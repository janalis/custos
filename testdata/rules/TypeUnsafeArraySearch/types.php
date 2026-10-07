<?php
function a1() {
    /** @var integer $a */
    /** @var integer[] $b */
    return in_array($a, $b);
}
function a2() {
    /** @var \int $a */
    /** @var \int[] $b */
    return in_array($a, $b);
}
function a3() {
    /** @var string $a */
    /** @var array<int, string> $b */
    return in_array($a, $b);
}
function a4() {
    /** @var int $a */
    /** @var array<int> $b */
    return in_array($a, $b);
}
function a5() {
    /** @var \stdClass $a */
    /** @var \stdClass[] $b */
    return in_array($a, $b);
}
function a6(int $a, array $b) {
    /** @var int[] $b */
    return in_array($a, $b);
}
/**
 * @param int $a
 * @param int[] $b
 */
function a7($a, $b) {
    return in_array($a, $b);
}
function a8() {
    $a = 1; $b = [1, 2];
    return array_search($a, $b);
}
function a9() {
    return <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('x', ['1', 2])</weak_warning>;
}
