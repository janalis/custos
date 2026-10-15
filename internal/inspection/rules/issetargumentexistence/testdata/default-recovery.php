<?php
// The parser recovers an existence check in a constant-default position.
// It is outside the local body, so it cannot prove an undefined body local.
function recoveredDefault($default = isset($missing)) {}
