package metrics

// ApplicationOption используется для конфигурации Application
type ApplicationOption func(*Application)

// WithSnapshotter устанавливает snapshotter для Application
func WithSnapshotter(snapshotter snapshotter) ApplicationOption {
	return func(app *Application) {
		app.snapshotter = snapshotter
	}
}
