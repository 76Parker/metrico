package metrics

// ApplicationOption используется для конфигурации Application
type ApplicationOption func(*Application)

// WithSnapshotter устанавливает snapshotter для Application
func WithSnapshotter(notifier changeNotifier) ApplicationOption {
	return func(app *Application) {
		app.metricsChangeNotifier = notifier
	}
}
