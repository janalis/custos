<?php
function first_case($subject) {
    switch ($subject) {
        case <error descr="Use fmod for fractional operands.">2.5 % 2</error>: return 1;
        case <error descr="Use fmod for fractional operands.">3.5 % 2</error>: break;
    }
}
exit;
echo 4.5 % 2;
