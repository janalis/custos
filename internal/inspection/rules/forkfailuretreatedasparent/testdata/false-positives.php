<?php
function failureArmThenSpacedElseIf() {
    $pid = pcntl_fork();
    if ($pid === -1) {
        die('could not fork');
    } else if ($pid) {
        pcntl_wait($status);
    } else {
        exit(0);
    }
}
function reversedOperands() {
    $pid = pcntl_fork();
    if (-1 === $pid) { throw new RuntimeException('fork failed'); }
    if ($pid) { pcntl_wait($status); } else { exit(0); }
}
function statementBetweenGuardAndTest() {
    $pid = pcntl_fork();
    if ($pid === -1) { throw new RuntimeException('fork failed'); }
    $started = microtime(true);
    if ($pid) { pcntl_wait($status); } else { exit(0); }
}
function negativeRange() {
    $pid = pcntl_fork();
    if ($pid < 0) { throw new RuntimeException('fork failed'); }
    if ($pid) { pcntl_wait($status); } else { exit(0); }
}
function childHandledFirst() {
    $pid = pcntl_fork();
    if ($pid === -1) { return false; }
    if ($pid === 0) { exit(0); }
    if ($pid) { pcntl_wait($status); }
}
function looseAndMirroredBounds($verbose, array $jobs) {
    $pid = pcntl_fork();
    if ($verbose) {
        foreach ($jobs as $job) {
            if ($job === 1) {
                if ($pid <= -1) { return; }
                if (!$pid) { echo $job; }
            }
            if ($job === 2) {
                if (0 > $pid) { return; }
                if (!$pid) { echo $job; }
            }
            if (-1 >= $pid) { continue; }
            if (!$pid) { echo $job; }
        }
    }
}
function laterArmsOfOneChain($skip) {
    $pid = pcntl_fork();
    if ($skip) {
        echo 'skipped';
    } elseif ($pid == -1) {
        echo 'failed';
    } else {
        if ($pid) { echo 'parent'; }
    }
}
function guardOnTheStoringAssignment() {
    if (($pid = pcntl_fork()) === -1) { return; }
    if ($pid) { echo 'parent'; }
}
function storingAssignmentInTheChain() {
    if (($pid = pcntl_fork()) < 0) {
        echo 'failed';
    } else if (!$pid) {
        echo 'child';
    }
}
function unbracedBodies($verbose) {
    $pid = pcntl_fork();
    if ($pid === -1) return;
    if ($verbose) if ($pid) echo 'parent';
}
function boundHeldByLocal() {
    $pid = pcntl_fork();
    $failed = -1;
    if ($failed === $pid) { return; }
    if ($pid) { echo 'parent'; }
}
function guardBeforeEnclosingStatement($verbose) {
    $pid = pcntl_fork();
    if ($pid === -1) { return; }
    if ($verbose) {
        echo 'forked';
        if ($pid) { echo 'parent'; }
    }
}
function guardAroundSwitchAndTry($mode) {
    $pid = pcntl_fork();
    if ($pid === -1) { return; }
    switch ($mode) {
        case 1:
            try {
                if ($pid) { echo 'parent'; }
            } catch (Exception $e) {
                if (!$pid) { echo 'child'; }
            } finally {
                if ($pid) { echo 'parent'; }
            }
    }
}
