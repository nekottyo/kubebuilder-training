package controllers

import (
	"context"
	"time"

	viewv1 "github.com/nekottyo/kubebuilder-training/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MarkdownView controller", func() {
	//! [setup]
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &viewv1.MarkdownView{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		svcs := &corev1.ServiceList{}
		err = k8sClient.List(ctx, svcs, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		for _, svc := range svcs.Items {
			err := k8sClient.Delete(ctx, &svc)
			Expect(err).NotTo(HaveOccurred())
		}
		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		reconciler := MarkdownViewReconciler{
			Client: k8sClient,
			Scheme: scheme,
			// Recorder: mgr.GetEventRecorderFor("markdownview-controller"),
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	It("should create ConfigMap", func() {
		mdView := newMarkdownView()
		err := k8sClient.Create(ctx, mdView)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "markdowns-sample"}, &cm)
		}).Should(Succeed())
		Expect(cm.Data).Should(HaveKey("SUMMARY.md"))
		Expect(cm.Data).Should(HaveKey("page1.md"))
	})
})

func newMarkdownView() *viewv1.MarkdownView {
	return &viewv1.MarkdownView{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: viewv1.MarkdownViewSpec{
			Markdowns: map[string]string{
				"SUMMARY.md": `summary`,
				"page1.md":   `page1`,
			},
			Replicas:    3,
			ViewerImage: "peaceiris/mdbook:0.4.10",
		},
	}
}
