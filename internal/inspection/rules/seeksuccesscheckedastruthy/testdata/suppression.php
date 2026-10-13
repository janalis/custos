<?php
// @custos-ignore SeekSuccessCheckedAsTruthy
if (fseek($fp, 0)) { echo "seeked"; }
