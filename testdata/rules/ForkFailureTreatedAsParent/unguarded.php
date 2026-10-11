<?php
function guardDoesNotLeave() {
    $pid = pcntl_fork();
    if ($pid === -1) { echo 'failed'; }
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function guardOnAnotherCall() {
    $pid = pcntl_fork();
    if (pcntl_fork() === -1) { return; }
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function insideTheFailureArm() {
    $pid = pcntl_fork();
    if ($pid === -1) {
        if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
    }
}
function guardOnEarlierValue() {
    $pid = pcntl_fork();
    if ($pid === -1) { return; }
    $pid = pcntl_fork();
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function guardAfterTheTest() {
    $pid = pcntl_fork();
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
    if ($pid === -1) { return; }
}
function failureTestedInLaterArm($skip, $other) {
    $pid = pcntl_fork();
    if ($skip) {
        echo 'skipped';
    } elseif ($other) {
        if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
    } elseif ($pid === -1) {
        echo 'failed';
    }
}
function testsThatKeepMinusOne($unknown) {
    $pid = pcntl_fork();
    if ($pid > -1) { return; }
    if ($pid >= 0) { return; }
    if ($pid < -1) { return; }
    if ($pid <= -2) { return; }
    if ($pid === 0) { return; }
    if ($pid !== -1) { return; }
    if ($pid === $unknown) { return; }
    if ($pid + 1) { return; }
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function guardInsideClosure() {
    $pid = pcntl_fork();
    $check = function () use ($pid) { if ($pid === -1) { return; } };
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function compoundAssignmentGuard() {
    $total = 1;
    if (($total += pcntl_fork()) === -1) { return; }
    $pid = pcntl_fork();
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
function assignmentOfAnotherValue() {
    $pid = pcntl_fork();
    if (($other = lookup()) === -1) { return; }
    if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo 'parent'; }
}
