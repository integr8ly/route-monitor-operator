package adder_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	// tested package
	"github.com/openshift/route-monitor-operator/controllers/routemonitor/adder"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openshift/route-monitor-operator/api/v1alpha1"
	routemonitorconst "github.com/openshift/route-monitor-operator/pkg/const"
	consterror "github.com/openshift/route-monitor-operator/pkg/const/test/error"
	constinit "github.com/openshift/route-monitor-operator/pkg/const/test/init"
	customerrors "github.com/openshift/route-monitor-operator/pkg/util/errors"
	utilreconcile "github.com/openshift/route-monitor-operator/pkg/util/reconcile"
	clientmocks "github.com/openshift/route-monitor-operator/pkg/util/test/generated/mocks/client"
	"github.com/openshift/route-monitor-operator/pkg/util/test/helper"
	testhelper "github.com/openshift/route-monitor-operator/pkg/util/test/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/route-monitor-operator/controllers/routemonitor"
)

var _ = Describe("Adder", func() {
	var (
		scheme = constinit.Scheme
	)
	var (
		mockClient *clientmocks.MockClient
		mockCtrl   *gomock.Controller

		routeMonitorAdder       adder.RouteMonitorAdder
		routeMonitorAdderClient client.Client

		ctx context.Context

		routeMonitor           v1alpha1.RouteMonitor
		routeMonitorStatus     v1alpha1.RouteMonitorStatus
		routeMonitorFinalizers []string

		get    testhelper.MockHelper
		delete testhelper.MockHelper
		create testhelper.MockHelper
		update testhelper.MockHelper
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = clientmocks.NewMockClient(mockCtrl)

		routeMonitorAdderClient = mockClient

		ctx = constinit.Context

		get = testhelper.MockHelper{}
		delete = testhelper.MockHelper{}
		create = testhelper.MockHelper{}
		update = testhelper.MockHelper{}

		routeMonitorAdderClient = mockClient
		routeMonitorStatus = v1alpha1.RouteMonitorStatus{
			RouteURL: "fake-route-url",
		}
		routeMonitorFinalizers = routemonitorconst.FinalizerList

	})
	JustBeforeEach(func() {
		routeMonitorAdder = adder.RouteMonitorAdder{
			Log:    constinit.Logger,
			Client: routeMonitorAdderClient,
			Scheme: constinit.Scheme,
		}

		mockClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(update.ErrorResponse).
			Times(update.CalledTimes)

		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(get.ErrorResponse).
			Times(get.CalledTimes)

		mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).
			Return(delete.ErrorResponse).
			Times(delete.CalledTimes)

		mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).
			Return(create.ErrorResponse).
			Times(create.CalledTimes)

		routeMonitor = v1alpha1.RouteMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: routeMonitorFinalizers,
			},
			Status: routeMonitorStatus,
		}
	})
	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing CreateResourceIfNotFound", func() {
		BeforeEach(func() {
			// Arrange
			get = helper.NotFoundErrorHappensOnce()
			create.CalledTimes = 1
		})
		When("the resource Exists", func() {
			BeforeEach(func() {
				// Arrange
				get.ErrorResponse = nil
				create.CalledTimes = 0
			})
			It("should call `Get` and not call `Create`", func() {
				//Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				//Assert
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("the resource Get fails unexpectedly", func() {
			// Arrange
			BeforeEach(func() {
				get.ErrorResponse = consterror.CustomError
				create.CalledTimes = 0
			})
			It("should return the error and not call `Create`", func() {
				//Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})

		When("the resource is Not Found", func() {
			// Arrange
			It("should call `Get` successfully and `Create` the resource", func() {
				//Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				//Assert
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the resource Create fails unexpectedly", func() {
			BeforeEach(func() {
				// Arrange
				create.ErrorResponse = consterror.CustomError
			})
			It("should call `Get` Successfully and call `Create` but return the error", func() {
				//Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})

	})
	Describe("CreateBlackBoxExporterDeployment", func() {
		BeforeEach(func() {
			routeMonitorAdderClient = mockClient
			// Arrange
			get.CalledTimes = 1

		})

		When("the resource(deployment) Exists", func() {
			It("should call `Get` and not call `Create`", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterDeploymentExists(ctx)
				//Assert
				Expect(err).NotTo(HaveOccurred())

			})
		})
		When("the resource(deployment) is Not Found", func() {
			// Arrange
			BeforeEach(func() {
				get.ErrorResponse = consterror.NotFoundErr
				create.CalledTimes = 1
			})
			It("should call `Get` successfully and `Create` the resource(deployment)", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterDeploymentExists(ctx)
				//Assert
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("the resource(deployment) Get fails unexpectedly", func() {
			// Arrange
			BeforeEach(func() {
				get.ErrorResponse = consterror.CustomError
			})
			It("should return the error and not call `Create`", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterDeploymentExists(ctx)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})
		When("the resource(deployment) Create fails unexpectedly", func() {
			// Arrange
			BeforeEach(func() {
				get.ErrorResponse = consterror.NotFoundErr
				create = helper.CustomErrorHappensOnce()
			})
			It("should call `Get` Successfully and call `Create` but return the error", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterDeploymentExists(ctx)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})
	})
	Describe("CreateBlackBoxExporterService", func() {
		BeforeEach(func() {
			routeMonitorAdderClient = mockClient
		})

		When("the resource(service) Exists", func() {
			// Arrange
			BeforeEach(func() {
				get.CalledTimes = 1
			})
			It("should call `Get` and not call `Create`", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterServiceExists(ctx)
				//Assert
				Expect(err).NotTo(HaveOccurred())

			})
		})
		When("the resource(service) is Not Found", func() {
			// Arrange
			BeforeEach(func() {
				get = helper.NotFoundErrorHappensOnce()
				create.CalledTimes = 1
			})
			It("should call `Get` successfully and `Create` the resource(service)", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterServiceExists(ctx)
				//Assert
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("the resource(service) Get fails unexpectedly", func() {
			// Arrange
			BeforeEach(func() {
				get = helper.CustomErrorHappensOnce()
			})
			It("should return the error and not call `Create`", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterServiceExists(ctx)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})
		When("the resource(service) Create fails unexpectedly", func() {
			// Arrange
			BeforeEach(func() {
				get = helper.NotFoundErrorHappensOnce()
				create = helper.CustomErrorHappensOnce()
			})
			It("should call `Get` Successfully and call `Create` but return the error", func() {
				//Act
				err := routeMonitorAdder.EnsureBlackBoxExporterServiceExists(ctx)
				//Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})
	})
	Describe("CreateServiceMonitorResource", func() {
		When("the RouteMonitor has no Host", func() {
			// Arrange
			BeforeEach(func() {
				routeMonitorAdderClient = fake.NewFakeClientWithScheme(scheme)
				routeMonitorStatus = v1alpha1.RouteMonitorStatus{}
			})
			It("should return No Host error", func() {
				// Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(customerrors.NoHost))
			})
		})
		When("func 'Update' failed unexpectidly", func() {
			// Arrange
			BeforeEach(func() {
				routeMonitorAdderClient = mockClient
				update = helper.CustomErrorHappensOnce()
				routeMonitorFinalizers = nil
				routeMonitorStatus = v1alpha1.RouteMonitorStatus{
					RouteURL: "fake-route-url",
				}
			})
			It("should bubble up the error", func() {
				// Act
				_, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(consterror.CustomError))
			})
		})
		When("the RouteMonitor has no Finalizer", func() {
			// Arrange
			BeforeEach(func() {
				routeMonitorAdderClient = mockClient
				update.CalledTimes = 1
				routeMonitorFinalizers = nil
				routeMonitorStatus = v1alpha1.RouteMonitorStatus{
					RouteURL: "fake-route-url",
				}
			})
			It("Should update the RouteMonitor with the finalizer", func() {
				// Act
				resp, err := routeMonitorAdder.EnsureServiceMonitorResourceExists(ctx, routeMonitor)
				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp).To(Equal(utilreconcile.StopOperation()))
			})
		})
	})

	Describe("New", func() {
		When("func New is called", func() {
			It("should return a new Deleter object", func() {
				// Arrange
				r := routemonitor.RouteMonitorReconciler{
					Client: routeMonitorAdderClient,
					Log:    constinit.Logger,
					Scheme: constinit.Scheme,
				}
				// Act
				res := adder.New(r)
				// Assert
				Expect(res).To(Equal(&adder.RouteMonitorAdder{
					Client: routeMonitorAdderClient,
					Log:    constinit.Logger,
					Scheme: constinit.Scheme,
				}))
			})
		})
	})
})
