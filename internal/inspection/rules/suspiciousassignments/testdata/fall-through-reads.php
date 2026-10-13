<?php
function runSteps(int $from): int {
    switch ($from) {
        case 1:
            $errors = checkFirst();
            if ($errors !== []) {
                return 1;
            }
            // no break
        case 2:
            $errors = checkSecond();
            if ($errors !== []) {
                return 2;
            }
            // no break
        case 3:
            $note = 'three';
            $errors = checkThird();
            report($errors);
        case 4:
            $errors = checkFourth();
            <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$note</error> = 'four';
            [$errors, $extra] = pair();
            return count($errors) + $extra;
    }
    return 0;
}
function legacyShim(string $a, ?string $b = null): string {
    $b = \func_num_args() > 1 ? \func_get_arg(1) : 'none';
    return $a . $b;
}
function legacyAll(string $a, $b = null): array {
    $b = 'x';
    return \func_get_args();
}
function nestedOnly(string $a, $b = null) {
    <error descr="Parameter is overwritten before its value is used.">$b</error> = 'x';
    return static function () use ($b) { return \func_get_args(); };
}
