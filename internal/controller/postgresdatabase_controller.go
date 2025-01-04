/*
Copyright 2025.

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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"database/sql"

	_ "github.com/lib/pq"

	postgresdbmanagementv1alpha1 "github.com/dioniseo/kubernetes-pg-manager-operator/api/v1alpha1"
)

// PostgresDatabaseReconciler reconciles a PostgresDatabase object
type PostgresDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgres.db.management.minikube.local,resources=postgresdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.db.management.minikube.local,resources=postgresdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgres.db.management.minikube.local,resources=postgresdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgresDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *PostgresDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	postgresDbCr := &postgresdbmanagementv1alpha1.PostgresDatabase{}
	if err := r.Get(ctx, req.NamespacedName, postgresDbCr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Fetch the custom resource
	dbCredentialsSecretObject := &corev1.Secret{}
	dbCredentialsSecret := types.NamespacedName{
		Namespace: postgresDbCr.GetNamespace(),
		Name:      postgresDbCr.Spec.DbCredentialsSecret.SecretName,
	}

	if err := r.Get(ctx, dbCredentialsSecret, dbCredentialsSecretObject); err != nil {
		return ctrl.Result{}, err
	}

	username := string(dbCredentialsSecretObject.Data[postgresDbCr.Spec.DbCredentialsSecret.UsernameKey])
	password := string(dbCredentialsSecretObject.Data[postgresDbCr.Spec.DbCredentialsSecret.PasswordKey])

	dbHost := postgresDbCr.Spec.DbHost
	dbPort := postgresDbCr.Spec.DbPort
	dbName := postgresDbCr.Spec.DbName

	fmt.Printf("Creaing Database with the following parameters:\nHost: %s, Port: %d, DbName: %s, Username: %s, Password: %s\n",
		dbHost, dbPort, dbName, username, password)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, username, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to the database!")

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&postgresdbmanagementv1alpha1.PostgresDatabase{}).
		Named("postgresdatabase").
		Complete(r)
}
