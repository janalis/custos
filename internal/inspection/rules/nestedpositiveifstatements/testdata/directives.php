<?php
function directives($a, $b, $c)
{
    // A directive for the next line would act on another one once moved.
    if ($a) {
        // @phpstan-ignore-next-line
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($b->maybe()) {
            go();
        }
    }
    if ($a) {
        go();
    } else {
        /** @noinspection PhpUndefinedMethodInspection */
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go();
        }
    }
    if ($a) {
        // A plain note moves with the body.
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c) {
            go();
        }
    }
}
