#!/bin/bash

# 创建主要目录
mkdir -p api/{openapi,proto}
mkdir -p build/{docker,package}
mkdir -p cmd/{i18n-apiserver,i18n-authz-server}
mkdir -p configs
mkdir -p docs
mkdir -p internal/{apiserver,authzserver,pkg}
mkdir -p internal/apiserver/{config,controller,middleware,model,repository,service}
mkdir -p internal/authzserver/{config,controller,middleware,model,repository,service}
mkdir -p internal/pkg/{code,middleware,utils}
mkdir -p pkg
mkdir -p scripts
mkdir -p test
