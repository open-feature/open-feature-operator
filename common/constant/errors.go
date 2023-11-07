package constant

import "errors"

var ErrFlagdProxyNotReady = errors.New("flagd-proxy is not ready, deferring pod admission")
var ErrUnrecognizedSyncProvider = errors.New("unrecognized sync provider")
