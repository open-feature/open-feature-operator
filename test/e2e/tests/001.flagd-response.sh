#!/bin/bash

./"$(dirname "${BASH_SOURCE[0]}")"/../flag-evaluation.sh simple-flag '{"value":true,"reason":"STATIC","variant":"on"}'
./"$(dirname "${BASH_SOURCE[0]}")"/../flag-evaluation.sh simple-flag-filepath '{"value":true,"reason":"STATIC","variant":"on"}'
./"$(dirname "${BASH_SOURCE[0]}")"/../flag-evaluation.sh simple-flag-filepath2 '{"value":true,"reason":"STATIC","variant":"on"}'
