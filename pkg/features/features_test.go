package features_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
)

var _ = Describe("Features", func() {
	Describe("OptionSet", func() {
		Describe("NewOptionSet", func() {
			It("should create a new OptionSet with empty mounts", func() {
				opts := features.NewOptionSet()
				Expect(opts.Mounts).To(BeEmpty())
				Expect(opts.Mounts).NotTo(BeNil())
			})
		})

		Describe("AddVolumeMount", func() {
			It("should add a single volume mount", func() {
				opts := features.NewOptionSet()
				mount := engine.VolumeMount{
					Source:       "/source",
					Destination:  "/dest",
					MountOptions: "ro",
				}
				opts.AddVolumeMount(mount)
				Expect(opts.Mounts).To(HaveLen(1))
				Expect(opts.Mounts[0]).To(Equal(mount))
			})

			It("should add multiple volume mounts at once", func() {
				opts := features.NewOptionSet()
				mount1 := engine.VolumeMount{
					Source:       "/source1",
					Destination:  "/dest1",
					MountOptions: "ro",
				}
				mount2 := engine.VolumeMount{
					Source:       "/source2",
					Destination:  "/dest2",
					MountOptions: "rw",
				}
				opts.AddVolumeMount(mount1, mount2)
				Expect(opts.Mounts).To(HaveLen(2))
				Expect(opts.Mounts[0]).To(Equal(mount1))
				Expect(opts.Mounts[1]).To(Equal(mount2))
			})

			It("should append to existing mounts", func() {
				opts := features.NewOptionSet()
				mount1 := engine.VolumeMount{
					Source:       "/source1",
					Destination:  "/dest1",
					MountOptions: "ro",
				}
				mount2 := engine.VolumeMount{
					Source:       "/source2",
					Destination:  "/dest2",
					MountOptions: "rw",
				}
				opts.AddVolumeMount(mount1)
				opts.AddVolumeMount(mount2)
				Expect(opts.Mounts).To(HaveLen(2))
				Expect(opts.Mounts[0]).To(Equal(mount1))
				Expect(opts.Mounts[1]).To(Equal(mount2))
			})
		})
	})

	Describe("Register", func() {
		var mockFeature *MockFeature

		BeforeEach(func() {
			mockFeature = &MockFeature{}
		})

		It("should register a new feature", func() {
			err := features.Register("test-feature", mockFeature)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when registering duplicate feature", func() {
			err := features.Register("duplicate-feature", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			err = features.Register("duplicate-feature", mockFeature)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already registered"))
		})
	})

	Describe("Initialize", func() {
		BeforeEach(func() {
			features.Reset()
		})

		It("should initialize enabled features and collect their options", func() {
			mockFeature := &MockFeature{
				enabled: true,
				options: features.OptionSet{
					Mounts: []engine.VolumeMount{
						{Source: "/test", Destination: "/test", MountOptions: "ro"},
					},
				},
			}
			err := features.Register("init-test-1", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			opts, err := features.Initialize()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockFeature.configureCalled).To(BeTrue())
			Expect(mockFeature.initializeCalled).To(BeTrue())
			Expect(opts.Mounts).To(ContainElement(mockFeature.options.Mounts[0]))
		})

		It("should skip disabled features", func() {
			mockFeature := &MockFeature{
				enabled: false,
			}
			err := features.Register("init-test-disabled", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			_, err = features.Initialize()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockFeature.configureCalled).To(BeTrue())
			Expect(mockFeature.initializeCalled).To(BeFalse())
		})

		It("should skip features that fail configuration", func() {
			mockFeature := &MockFeature{
				configureError: Errorf("config error"),
			}
			err := features.Register("init-test-config-fail", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			_, err = features.Initialize()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockFeature.configureCalled).To(BeTrue())
			Expect(mockFeature.initializeCalled).To(BeFalse())
		})

		It("should call HandleError when initialization fails", func() {
			mockFeature := &MockFeature{
				enabled:        true,
				initializeError: Errorf("init error"),
				exitOnError:    false,
			}
			err := features.Register("init-test-init-fail", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			_, err = features.Initialize()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockFeature.handleErrorCalled).To(BeTrue())
		})

		It("should return error when feature exits on error", func() {
			mockFeature := &MockFeature{
				enabled:        true,
				initializeError: Errorf("fatal error"),
				exitOnError:    true,
			}
			err := features.Register("init-test-exit-on-error", mockFeature)
			Expect(err).NotTo(HaveOccurred())

			_, err = features.Initialize()
			Expect(err).To(HaveOccurred())
			Expect(mockFeature.handleErrorCalled).To(BeTrue())
		})

		It("should collect mounts from multiple features", func() {
			mockFeature1 := &MockFeature{
				enabled: true,
				options: features.OptionSet{
					Mounts: []engine.VolumeMount{
						{Source: "/test1", Destination: "/test1", MountOptions: "ro"},
					},
				},
			}
			mockFeature2 := &MockFeature{
				enabled: true,
				options: features.OptionSet{
					Mounts: []engine.VolumeMount{
						{Source: "/test2", Destination: "/test2", MountOptions: "rw"},
					},
				},
			}
			err := features.Register("init-test-multi-1", mockFeature1)
			Expect(err).NotTo(HaveOccurred())
			err = features.Register("init-test-multi-2", mockFeature2)
			Expect(err).NotTo(HaveOccurred())

			opts, err := features.Initialize()
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Mounts).To(ContainElement(mockFeature1.options.Mounts[0]))
			Expect(opts.Mounts).To(ContainElement(mockFeature2.options.Mounts[0]))
		})
	})
})

// MockFeature is a mock implementation of the Feature interface for testing
type MockFeature struct {
	enabled            bool
	configureError     error
	initializeError    error
	exitOnError        bool
	options            features.OptionSet
	configureCalled    bool
	initializeCalled   bool
	handleErrorCalled  bool
}

func (m *MockFeature) Configure() error {
	m.configureCalled = true
	return m.configureError
}

func (m *MockFeature) Initialize() (features.OptionSet, error) {
	m.initializeCalled = true
	return m.options, m.initializeError
}

func (m *MockFeature) Enabled() bool {
	return m.enabled
}

func (m *MockFeature) HandleError(err error) {
	m.handleErrorCalled = true
}

func (m *MockFeature) ExitOnError() bool {
	return m.exitOnError
}

func Errorf(format string, args ...interface{}) error {
	return &mockError{msg: format}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
