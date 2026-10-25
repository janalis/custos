<?php
function directives($a, $b, $c)
{
    // A directive for the next line would act on another one once moved.
    if ($a) {
        // @phpstan-ignore-next-line
        if ($b->maybe()) {
            go();
        }
    }
    if ($a) {
        go();
    } else {
        /** @noinspection PhpUndefinedMethodInspection */
        if ($c) {
            go();
        }
    }
    if ($a && $c) {
            // A plain note moves with the body.
            go();
        }
}
