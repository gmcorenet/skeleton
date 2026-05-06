module github.com/gmcorenet/skeleton

go 1.21

require (
	github.com/gmcorenet/framework v1.0.0
	github.com/gmcorenet/sdk/gmcore-debugbar v0.1.0
)

require gopkg.in/yaml.v3 v3.0.1

replace github.com/gmcorenet/sdk/gmcore-debugbar => ../sdks/gmcore-debugbar
