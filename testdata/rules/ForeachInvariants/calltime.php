<?php
function calltime($tokens) {
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: call-time by-reference argument
        echo $tokens[$k];
        unknownFn(&$k);
    }
}
