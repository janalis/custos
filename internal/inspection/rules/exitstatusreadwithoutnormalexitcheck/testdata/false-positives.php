<?php
function nestedIfs(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r > 0) {
        if (pcntl_wifexited($status)) {
            return pcntl_wexitstatus($status);
        }
    }
    return null;
}
function earlyReturns(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r <= 0) { return null; }
    if (!pcntl_wifexited($status)) { return null; }
    return pcntl_wexitstatus($status);
}
function combinedEarlyReturn(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r < 1 || !pcntl_wifexited($status)) {
        throw new RuntimeException('child did not exit');
    }
    foreach ([1] as $unused) {
        echo pcntl_wexitstatus($status);
    }
}
function ternaryThen(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    return ($r > 0 && pcntl_wifexited($status)) ? pcntl_wexitstatus($status) : null;
}
function ternaryElse(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    return ($r === -1 || !pcntl_wifexited($status)) ? null : pcntl_wexitstatus($status);
}
function mirrored(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if (0 < $r && pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    $r = pcntl_waitpid($pid, $status);
    if (-1 !== $r and pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    $r = pcntl_waitpid($pid, $status, WNOHANG);
    if ($r >= 1 && pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    $r = pcntl_wait($status);
    if (1 <= $r && pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    return null;
}
function elseBranches(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r == -1) {
        return null;
    } elseif (!pcntl_wifexited($status)) {
        return false;
    } else {
        return pcntl_wexitstatus($status);
    }
}
function elseIfBranch(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r <= 0) {
        return null;
    } elseif (pcntl_wifexited($status)) {
        return pcntl_wexitstatus($status);
    }
    return false;
}
function sameCondition(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    $clean = $r > 0 && pcntl_wifexited($status) && pcntl_wexitstatus($status) === 0;
    $failed = ($r <= 0 or !pcntl_wifexited($status) or pcntl_wexitstatus($status) !== 0);
    $other = !($r > 0 && pcntl_wifexited($status)) ? null : pcntl_wexitstatus($status);
    return $clean || $failed;
}
function directResult(int $pid) {
    if (pcntl_waitpid($pid, $status) > 0 && pcntl_wifexited($status)) {
        return pcntl_wexitstatus($status);
    }
    return null;
}
function blockingBounds(int $pid) {
    // A blocking wait returns a child pid or -1: excluding -1 is enough.
    $r = pcntl_waitpid($pid, $status);
    if ($r >= 0 && pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    $r = pcntl_waitpid($pid, $status);
    if ($r > -1 && pcntl_wifexited($status)) { return pcntl_wexitstatus($status); }
    return null;
}
