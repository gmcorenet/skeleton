module github.com/gmcorenet/skeleton

go 1.23

require (
	github.com/gmcorenet/framework v1.0.0
	github.com/gmcorenet/sdk-gmcore-asset-mapper v1.0.0
	github.com/gmcorenet/sdk-gmcore-debugbar v0.1.0
	github.com/gmcorenet/sdk-gmcore-error v1.0.0
	github.com/gmcorenet/sdk-gmcore-filesystem v0.1.0
	github.com/gmcorenet/sdk-gmcore-form v0.1.0
	github.com/gmcorenet/sdk-gmcore-httpclient v0.1.0
	github.com/gmcorenet/sdk-gmcore-i18n v0.1.0
	github.com/gmcorenet/sdk-gmcore-log v0.1.0
	github.com/gmcorenet/sdk-gmcore-mailer v0.1.0
	github.com/gmcorenet/sdk-gmcore-messenger v0.1.0
	github.com/gmcorenet/sdk-gmcore-migrations v0.0.0-20260507103628-9dbf0cc44249
	github.com/gmcorenet/sdk-gmcore-scheduler v0.1.0
	github.com/gmcorenet/sdk-gmcore-serializer v0.1.0
	github.com/gmcorenet/sdk-gmcore-templating v0.1.0
	github.com/gmcorenet/sdk-gmcore-transport v0.1.0
	github.com/gmcorenet/sdk-gmcore-webhook v0.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gmcorenet/sdk-gmcore-config v1.0.0 // indirect
	github.com/gmcorenet/sdk-gmcore-events v0.1.0 // indirect
	github.com/gmcorenet/sdk-gmcore-validation v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	github.com/gmcorenet/framework => ../framework
	github.com/gmcorenet/sdk-gmcore-asset-mapper => ../sdks/gmcore-asset-mapper
	github.com/gmcorenet/sdk-gmcore-config => ../sdks/gmcore-config
	github.com/gmcorenet/sdk-gmcore-debugbar => ../sdks/gmcore-debugbar
	github.com/gmcorenet/sdk-gmcore-error => ../sdks/gmcore-error
	github.com/gmcorenet/sdk-gmcore-events => ../sdks/gmcore-events
	github.com/gmcorenet/sdk-gmcore-filesystem => ../sdks/gmcore-filesystem
	github.com/gmcorenet/sdk-gmcore-form => ../sdks/gmcore-form
	github.com/gmcorenet/sdk-gmcore-httpclient => ../sdks/gmcore-httpclient
	github.com/gmcorenet/sdk-gmcore-i18n => ../sdks/gmcore-i18n
	github.com/gmcorenet/sdk-gmcore-log => ../sdks/gmcore-log
	github.com/gmcorenet/sdk-gmcore-mailer => ../sdks/gmcore-mailer
	github.com/gmcorenet/sdk-gmcore-messenger => ../sdks/gmcore-messenger
	github.com/gmcorenet/sdk-gmcore-scheduler => ../sdks/gmcore-scheduler
	github.com/gmcorenet/sdk-gmcore-serializer => ../sdks/gmcore-serializer
	github.com/gmcorenet/sdk-gmcore-templating => ../sdks/gmcore-templating
	github.com/gmcorenet/sdk-gmcore-transport => ../sdks/gmcore-transport
	github.com/gmcorenet/sdk-gmcore-validation => ../sdks/gmcore-validation
	github.com/gmcorenet/sdk-gmcore-webhook => ../sdks/gmcore-webhook
)
