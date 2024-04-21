/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"os/exec"
	"path/filepath"

	//"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secopsv1alpha1 "github.com/kavinduxo/sentinel-operator/api/v1alpha1"
)

const sentinelFinalizer = "secops.kavinduxo.com/finalizer"

// Definitions to manage status conditions
const (
	// typeAvailableSentinel represents the status of the Deployment reconciliation
	typeAvailableSentinel = "Available"
	// typeDegradedSentinel represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	typeDegradedSentinel     = "Degraded"
	typeRbacIssueSentinel    = "RBAC-Failed"
	typeEncryptIssueSentinel = "Encryption-Failed"
)

const (
	typeSecretBase              = "BaseSecret"
	typeSecretBaseRbac          = "RbacBaseSecret"
	typeSecretLocalEncryted     = "SecuredSecret"
	typeSecretLocalEncrytedRbac = "RbacSecuredSecret"
	typeSecretKmsEncrypted      = "KMSSecuredSecret"
	typeSecretKmsEncryptedRbac  = "RbacKMSSecuredSecret"
)

// SentinelReconciler reconciles a Sentinel object
type SentinelReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=secops.kavinduxo.com,resources=sentinels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secops.kavinduxo.com,resources=sentinels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secops.kavinduxo.com,resources=sentinels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sentinel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *SentinelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Sentinel instance
	// The purpose is check if the Custom Resource for the Kind Sentinel
	// is applied on the cluster if not we return nil to stop the reconciliation
	sentinel := &secopsv1alpha1.Sentinel{}
	err := r.Get(ctx, req.NamespacedName, sentinel)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("sentinel resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get sentinel")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if sentinel.Status.Conditions == nil || len(sentinel.Status.Conditions) == 0 {
		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeAvailableSentinel, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, sentinel); err != nil {
			log.Error(err, "Failed to update Sentinel status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the sentinel Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raise the issue "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, sentinel); err != nil {
			log.Error(err, "Failed to re-fetch sentinel")
			return ctrl.Result{}, err
		}
	}

	// Let's add a finalizer. Then, we can define some operations which should
	// occurs before the custom resource to be deleted.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(sentinel, sentinelFinalizer) {
		log.Info("Adding Finalizer for Sentinel")
		if ok := controllerutil.AddFinalizer(sentinel, sentinelFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, sentinel); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the Sentinel instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isSentinelMarkedToBeDeleted := sentinel.GetDeletionTimestamp() != nil
	if isSentinelMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(sentinel, sentinelFinalizer) {
			log.Info("Performing Finalizer Operations for Sentinel before delete CR")

			// Let's add here an status "Downgrade" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeDegradedSentinel,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", sentinel.Name)})

			if err := r.Status().Update(ctx, sentinel); err != nil {
				log.Error(err, "Failed to update Sentinel status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before remove the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			r.doFinalizerOperationsForSentinel(sentinel)

			// TODO(user): If you add operations to the doFinalizerOperationsForSentinel method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the sentinel Custom Resource before update the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, sentinel); err != nil {
				log.Error(err, "Failed to re-fetch sentinel")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeDegradedSentinel,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", sentinel.Name)})

			if err := r.Status().Update(ctx, sentinel); err != nil {
				log.Error(err, "Failed to update Sentinel status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for Sentinel after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(sentinel, sentinelFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for Sentinel")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, sentinel); err != nil {
				log.Error(err, "Failed to remove finalizer for Sentinel")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	secret, secretForSentinelRes, err := r.secretForSentinel(sentinel, ctx, req)
	if err != nil {
		return secretForSentinelRes, err
	}

	log.Info("Secret is Available now",
		"Secret.Namespace", secret.Namespace, "Seret.Name", secret.Name)

	// Secret created successfully
	// We will requeue the reconciliation so that we can ensure the state
	// and move forward for the next operations

	// Validate the custom resource spec
	if validateRes, err := r.validateSentinelSpec(sentinel, ctx, req); err != nil {
		return validateRes, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{
		Type:   typeAvailableSentinel,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Secret for custom resource created successfully (%s): (%s) : (%s)", sentinel.Name, sentinel.Namespace, sentinel.Spec.SecretType),
	})

	if err := r.Status().Update(ctx, sentinel); err != nil {
		log.Error(err, "Failed to update Sentinel status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

func (r *SentinelReconciler) validateSentinelSpec(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	crName := sentinel.ObjectMeta.Name
	crKind := sentinel.Kind

	if crKind != "Sentinel" {
		kindErr := errors.New(crKind + " is an invalid Kind for the Sentinel CR.")
		log.Error(kindErr, "Invalid Kind!")

		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeAvailableSentinel,
			Status: metav1.ConditionFalse, Reason: "Resizing",
			Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", sentinel.Name, kindErr)})

		return ctrl.Result{}, kindErr
	}

	if crName == "" {
		crNameErr := errors.New("Sentinel CR didn't map a name to metadata")
		log.Error(crNameErr, "Invalid Metadata!")

		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeAvailableSentinel,
			Status: metav1.ConditionFalse, Reason: "Resizing",
			Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", sentinel.Name, crNameErr)})

		return ctrl.Result{}, crNameErr
	}

	if crTypeErr := r.validateSecretType(sentinel); crTypeErr != nil {
		log.Error(crTypeErr, "Invalid Secret Type!")

		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeAvailableSentinel,
			Status: metav1.ConditionFalse, Reason: "Resizing",
			Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", sentinel.Name, crTypeErr)})

		return ctrl.Result{}, crTypeErr
	}

	// Now, that we update the size we want to requeue the reconciliation
	// so that we can ensure that we have the latest state of the resource before
	// update. Also, it will help ensure the desired state on the cluster
	return ctrl.Result{Requeue: true}, nil

}

func (r *SentinelReconciler) validateSecretType(
	sentinel *secopsv1alpha1.Sentinel) error {
	if sentinel.Spec.SecretType == "" {
		return fmt.Errorf("SecretType is required")
	}

	// Check if SecretType is one of the allowed values
	switch sentinel.Spec.SecretType {
	case typeSecretBase, typeSecretBaseRbac, typeSecretLocalEncryted, typeSecretLocalEncrytedRbac, typeSecretKmsEncrypted, typeSecretKmsEncryptedRbac:
		// Valid SecretType
		return nil
	default:
		// Invalid SecretType
		return fmt.Errorf("Invalid SecretType: %s", sentinel.Spec.SecretType)
	}
}

// finalizeSentinel will perform the required operations before delete the CR.
func (r *SentinelReconciler) doFinalizerOperationsForSentinel(cr *secopsv1alpha1.Sentinel) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers with the purpose of delete resources which are
	// created and managed in the reconciliation. These ones, such as the Deployment created on this reconcile,
	// are defined as depended of the custom resource. See that we use the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/

	// The following implementation will raise an event
	// r.Recorder.Event(cr, "Warning", "Deleting",
	// 	fmt.Sprintf("Custom Resource %s is being deleted from the namespace %s",
	// 		cr.Name,
	// 		cr.Namespace))
}

func (r *SentinelReconciler) secretForSentinel(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request) (*corev1.Secret, ctrl.Result, error) {

	log := log.FromContext(ctx)
	secretName := sentinel.Spec.SecretName
	secretNamespace := sentinel.Namespace
	secretType := sentinel.Spec.SecretType

	// Fetch the Secret if it exists
	existSecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, existSecret)

	if err != nil && apierrors.IsNotFound(err) {
		// Secret does not exist, create a new one

		// Create the Secret data
		secretData := map[string][]byte{}
		for key, value := range sentinel.Spec.Data {
			//Check the type of the secret
			if secretType == typeSecretBaseRbac {
				if validateRbacSecretRes, err := r.validateRbacSecret(sentinel, ctx, req); err != nil {
					return nil, validateRbacSecretRes, err
				}
				secretData[key] = []byte(value)

			} else if secretType == typeSecretLocalEncryted {
				validateLocalEncryptedSecretRes, err, ecryptedSecretData := r.validateLocalEncryptedSecret2(sentinel, ctx, req, value)
				if err != nil {
					return nil, validateLocalEncryptedSecretRes, err
				}
				secretData[key] = ecryptedSecretData

			} else if secretType == typeSecretLocalEncrytedRbac {
				validateLocalEncryptedSecretRes, err, ecryptedSecretData := r.validateLocalEncryptedSecret2(sentinel, ctx, req, value)
				if err != nil {
					return nil, validateLocalEncryptedSecretRes, err
				}
				if validateRbacSecretRes, err := r.validateRbacSecret(sentinel, ctx, req); err != nil {
					return nil, validateRbacSecretRes, err
				}
				secretData[key] = ecryptedSecretData
			} else {
				secretData[key] = []byte(value)
			}
		}

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: secretNamespace,
			},
			Data: secretData,
			Type: corev1.SecretTypeOpaque,
		}

		// Set Sentinel instance as the owner of the Secret
		if err := controllerutil.SetControllerReference(sentinel, newSecret, r.Scheme); err != nil {
			log.Error(err, "Setting Sentinel instance as the owner of the Secret Failed.")
			return nil, ctrl.Result{}, err
		}

		// Create the Secret
		if err := r.Create(ctx, newSecret); err != nil {
			log.Error(err, "Secret Creation Final Step Failed.")
			return nil, ctrl.Result{}, err
		}

		return newSecret, ctrl.Result{}, nil
	} else if err != nil {
		//if there is any error while fetching the existing secret
		return nil, ctrl.Result{}, err
	}

	return existSecret, ctrl.Result{}, nil
}

func (r *SentinelReconciler) validateRbacSecret(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	inputServiceAccount := sentinel.Spec.ServiceAccount
	inputRole := sentinel.Spec.Role
	inputRoleBinding := sentinel.Spec.RoleBinding
	inputNamespace := sentinel.Namespace
	inputUserType := sentinel.ObjectMeta.GetLabels()["usertype"]

	if inputRole == "" {
		inpRoErr := fmt.Errorf("Defining your Role is must under Spec.Role.")
		log.Error(inpRoErr, "Role Not Found!")

		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
			Status: metav1.ConditionFalse, Reason: "NotFound",
			Message: fmt.Sprintf("Role Not Found (%s): (%s)", sentinel.Name, inpRoErr)})

		return ctrl.Result{}, inpRoErr
	}

	if inputRoleBinding == "" {
		inpRbErr := fmt.Errorf("Defining your RoleBinding is must under Spec.RoleBinding.")
		log.Error(inpRbErr, "RoleBinding Not Found!")

		meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
			Status: metav1.ConditionFalse, Reason: "NotFound",
			Message: fmt.Sprintf("RoleBinding Not Found (%s): (%s)", sentinel.Name, inpRbErr)})

		return ctrl.Result{}, inpRbErr
	}

	if inputServiceAccount != "" {
		if inputUserType == "User" {
		} else if inputUserType == "ServiceAccount" {
			// Check if SA exists
			sa := &corev1.ServiceAccount{}
			saErr := r.Get(context.TODO(), types.NamespacedName{Name: inputServiceAccount, Namespace: inputNamespace}, sa)
			if saErr != nil {
				log.Error(saErr, "Service Account must create!")

				meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
					Status: metav1.ConditionFalse, Reason: "Not Found",
					Message: fmt.Sprintf("Service Account Not Found (%s): (%s)", sentinel.Name, saErr)})

				return ctrl.Result{}, saErr
			}
		} else {
			inpRbErr := fmt.Errorf("Define the User or Service Account in the labaels as an usertype.")
			log.Error(inpRbErr, "Invalid usertype!")

			meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
				Status: metav1.ConditionFalse, Reason: "Invalid",
				Message: fmt.Sprintf("usertype Not Found (%s): (%s)", sentinel.Name, inpRbErr)})

			return ctrl.Result{}, inpRbErr
		}
	}

	// Check if Role exists
	role := &rbacv1.Role{}
	roleErr := r.Get(context.TODO(), types.NamespacedName{Name: inputRole, Namespace: inputNamespace}, role)

	if roleErr != nil && apierrors.IsNotFound(roleErr) {

		// Role does not exist, create a new one
		newRole := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inputRole,
				Namespace: inputNamespace,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},        // Empty string represents the core API group
					Resources: []string{"secrets"}, // Resource types that this Role has access to
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		}

		// Set Sentinel instance as the owner of the Secret
		if err := controllerutil.SetControllerReference(sentinel, newRole, r.Scheme); err != nil {
			log.Error(err, "Setting Sentinel instance as the owner of the Role Failed.")
			return ctrl.Result{}, err
		}

		// Create the Role
		if err := r.Create(ctx, newRole); err != nil {
			log.Error(err, "Role Creation Final Step Failed.")
			return ctrl.Result{}, err
		}
	} else if roleErr != nil {
		//if there is any error while fetching the existing role
		return ctrl.Result{}, roleErr
	}

	roleBinding := &rbacv1.RoleBinding{}
	roleBindingErr := r.Get(context.TODO(), types.NamespacedName{Name: inputRoleBinding, Namespace: inputNamespace}, roleBinding)
	if roleBindingErr != nil && apierrors.IsNotFound(roleBindingErr) {

		// RoleBinding does not exist, create a new one
		newRole := &rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      inputRoleBinding,
				Namespace: inputNamespace,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     inputUserType,
					APIGroup: "",
					Name:     inputServiceAccount,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "",
				Kind:     "Role",
				Name:     inputRole,
			},
		}

		// Set Sentinel instance as the owner of the Secret
		if err := controllerutil.SetControllerReference(sentinel, newRole, r.Scheme); err != nil {
			log.Error(err, "Setting Sentinel instance as the owner of the Role Failed.")
			return ctrl.Result{}, err
		}

		// Create the Role
		if err := r.Create(ctx, newRole); err != nil {
			log.Error(err, "Role Creation Final Step Failed.")
			return ctrl.Result{}, err
		}
	} else if roleBindingErr != nil {
		//if there is any error while fetching the existing role binding
		return ctrl.Result{}, roleBindingErr
	}

	meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
		Status: metav1.ConditionFalse, Reason: "Binding",
		Message: fmt.Sprintf("Service Account, Role & Role Binding are good to go. (%s)", sentinel.Name)})

	return ctrl.Result{}, nil
}

// bcrypt secret securing (Start)
func (r *SentinelReconciler) validateLocalEncryptedSecret2(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request, secret string) (ctrl.Result, error, []byte) {

	log := log.FromContext(ctx)

	key := os.Getenv("ENCRYPTION_SECRET")
	if key == "" {
		keyErr := fmt.Errorf("Encryption key not found in .env file.")
		log.Error(keyErr, "KeyError")
	}

	encryptedSecret, err := encryptSecret(secret, []byte(key))
	if err != nil {
		log.Error(err, "Failed to encrypt secret")
	}

	meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
		Status: metav1.ConditionFalse, Reason: "Encrpt",
		Message: fmt.Sprintf("Encryption is locally done. (%s)", sentinel.Name)})

	return ctrl.Result{}, nil, encryptedSecret
}

func encryptSecret(secret string, key []byte) ([]byte, error) {
	// Hash secret using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}
	return hash, nil
}

//bcrypt secret securing (End)

// new way to handle secure secrets (start)
func (r *SentinelReconciler) validateLocalEncryptedSecret1(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	kubeApiServerConfig := "/etc/kubernetes/manifests/kube-apiserver.yaml"

	// this part of solution is based for minikube. it has to be changed according to the cluster architecture
	if err := sshIntoMinikube(); err != nil {
		log.Error(err, "Ping to control plane failed!")
	}

	enaled, err := isEncryptionProviderEnabled(kubeApiServerConfig)
	if err != nil {
		log.Error(err, "Couldn't find KubeApiServer config!")
	}

	if !enaled {
		if err := modifyKubeApiServer(); err != nil {
			log.Error(err, "Failed to modify KubeApiServer config!")
		}

		if err := createEncryptConfig(); err != nil {
			log.Error(err, "Failed to create Encryption Configuration file!")
		}

		if err := restartKubeApiServerContainer(); err != nil {
			log.Error(err, "Failed to restart the kube api server container!")
		}
	}

	meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
		Status: metav1.ConditionFalse, Reason: "Encrpt",
		Message: fmt.Sprintf("Encryption is locally setup correclty. (%s)", sentinel.Name)})

	return ctrl.Result{}, nil
}

func restartKubeApiServerContainer() error {
	cmd := exec.Command("sudo", "systemctl", "restart", "kubelet")
	return cmd.Run()
}

func createEncryptConfig() error {
	encryptionConfigContent := `
kind: EncryptionConfiguration
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: "` + os.Getenv("ENCRYPTION_SECRET") + `"
`
	file, err := os.Create("/etc/kubernetes/encryption-config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(encryptionConfigContent)
	if err != nil {
		return err
	}
	return nil
}

func modifyKubeApiServer() error {
	panic("unimplemented")
}

func isEncryptionProviderEnabled(kubeApiServerConfig string) (bool, error) {
	content, err := os.ReadFile(kubeApiServerConfig)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), "--encryption-provider-config"), nil
}

func sshIntoMinikube() error {
	cmd := exec.Command("minikube", "ssh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

//new way to handle secure secrets (end)

func (r *SentinelReconciler) validateLocalEncryptedSecret(
	sentinel *secopsv1alpha1.Sentinel, ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		log.Error(err, "Build Config Error!")
		return ctrl.Result{}, err
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Client Set Config Error!")
		return ctrl.Result{}, err
	}

	// Copy file to PKI location
	copyErr := copyFileToPKI("/home/kavindu/projects/operators/this/sentinel-operator/encryptconfig.yaml", clientset)
	if copyErr != nil {
		log.Error(err, "Error when copying a file to !")
		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(&sentinel.Status.Conditions, metav1.Condition{Type: typeRbacIssueSentinel,
		Status: metav1.ConditionFalse, Reason: "Encrpt",
		Message: fmt.Sprintf("Encryption is locally setup correclty. (%s)", sentinel.Name)})

	return ctrl.Result{}, nil
}

func copyFileToPKI(localFilePath string, clientset *kubernetes.Clientset) error {
	// Get the absolute path of the local file
	absPath, err := filepath.Abs(localFilePath)
	if err != nil {
		return err
	}

	// Open the local file
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	exists, err := fileExistsInPKI(clientset, filepath.Base(localFilePath))
	if err != nil {
		return err
	}
	if exists {
		log.Log.Info("File exists!")
		return nil
	}

	// Create a Kubernetes ConfigMap with the file data
	// You can also use Secrets if needed
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "file-configmap",
			Namespace: "kube-system", // Namespace where kube-controller-manager pod resides
		},
		Data: map[string]string{
			filepath.Base(absPath): absPath,
		},
	}

	// Create or update the ConfigMap in Kubernetes
	_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Use Kubernetes VolumeMount to mount the ConfigMap in kube-controller-manager pod
	// This requires modifying the pod spec and performing a rolling update

	return nil
}

func fileExistsInPKI(clientset *kubernetes.Clientset, fileName string) (bool, error) {
	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.Background(), "file-configmap", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	_, ok := cm.Data[fileName]
	return ok, nil
}

// labelsForSentinel returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForSentinel(name string) map[string]string {
	var imageTag string
	image, err := imageForSentinel()
	if err == nil {
		imageTag = strings.Split(image, ":")[1]
	}
	return map[string]string{"app.kubernetes.io/name": "Sentinel",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    imageTag,
		"app.kubernetes.io/part-of":    "sentinel-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

// imageForSentinel gets the Operand image which is managed by this controller
// from the SENTINEL_IMAGE environment variable defined in the config/manager/manager.yaml
func imageForSentinel() (string, error) {
	var imageEnvVar = "SENTINEL_IMAGE"
	image, found := os.LookupEnv(imageEnvVar)
	if !found {
		return "", fmt.Errorf("Unable to find %s environment variable with the image", imageEnvVar)
	}
	return image, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SentinelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secopsv1alpha1.Sentinel{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
