# Changelog

## v0.5.1

* update ffi

## v0.5.0

* bump uo version to v0.5.0
* update lotus version to v1.22.0
* update venus-auth version to v1.11.0
## v0.5.0-rc1

* Feat: recorver docker ci by @LinZexiao in https://github.com/ipfs-force-community/chain-co/pull/32
* fix: fix wrap rate litmit by @simlecode in https://github.com/ipfs-force-community/chain-co/pull/33
* feat: add status api to detect api ready by @hunjixin in https://github.com/ipfs-force-community/chain-co/pull/34
* fix: build deps before build docker by @LinZexiao in https://github.com/ipfs-force-community/chain-co/pull/36
* Feat: prevent launch without token by @LinZexiao in https://github.com/ipfs-force-community/chain-co/pull/41
* Feat: add docker push by @hunjixin in https://github.com/ipfs-force-community/chain-co/pull/42
* fix: block-not-found by @simlecode in https://github.com/ipfs-force-community/chain-co/pull/45
* chore: support MpoolBatchPushUntrusted by @simlecode in https://github.com/ipfs-force-community/chain-co/pull/37

## v0.4.0

* 升级 venus-auth 版本到 v1.10.0
* 升级 go-jsonrpc 版本到 v0.1.7
* 升级 lotus 版本到 v1.20.0

## v0.4.0-rc3

* 支持 MpoolBatchPushUntrusted [[#37](https://github.com/ipfs-force-community/chain-co/pull/37)]

## v0.4.0-rc2

* 升级 lotus 和 go-jsonrpc 版本
* 调整 Git workflows

## v0.4.0-rc1

支持 Filecoin NV18 网络升级

* 升级 lotus 版本到 v1.20.0-rc1
* 升级 venus-auth 版本到 v1.10.0-rc1

## v0.3.1

- 将 `ChainHead` 接口设置为本地 [[#28](https://github.com/ipfs-force-community/chain-co/pull/28)]
- 移除docker CI [[#27](https://github.com/ipfs-force-community/chain-co/pull/27)]

## v0.3.0

- 修复节点状态不一致导致区块找不到 [[#24](https://github.com/ipfs-force-community/chain-co/pull/24)]
