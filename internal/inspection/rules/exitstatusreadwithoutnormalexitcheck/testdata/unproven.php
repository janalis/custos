<?php
function samePid(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r === $pid && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    return null;
}
function staleCheck(int $pid, int $status) {
    if (pcntl_wifexited($status)) {
        $r = pcntl_waitpid($pid, $status);
        if ($r > 0) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    }
    return null;
}
function nonBlocking(int $pid) {
    $r = pcntl_waitpid($pid, $status, WNOHANG);
    if ($r !== -1 && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    $r = pcntl_wait($status, WNOHANG);
    if ($r != -1 && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    return null;
}
function weakBounds(int $pid) {
    $r = pcntl_waitpid($pid, $status, WNOHANG);
    if ($r >= 0 && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    $r = pcntl_waitpid($pid, $status, WNOHANG);
    if ($r > -1 && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    $r = pcntl_waitpid($pid, $status);
    if ($r !== 0 && pcntl_wifexited($status)) { return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>; }
    return null;
}
function wrongBranch(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r > 0 && pcntl_wifexited($status)) {
        return null;
    } elseif ($pid > 0) {
        return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    } else {
        return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    }
}
function wrongArm(int $pid, int $other) {
    $r = pcntl_waitpid($pid, $status);
    $a = ($r > 0 && pcntl_wifexited($status)) ? null : <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    $b = ($r <= 0 || !pcntl_wifexited($status)) ? <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning> : null;
    $c = $r > 0 && pcntl_wifexited($status) || <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    $d = ($r <= 0 || !pcntl_wifexited($status)) && <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    $e = $r > 0 && ~pcntl_wifexited($status) && $other > 0 && $pid > $other && <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    $f = !($r <= 0 || !pcntl_wifexited($status)) ? null : <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
    return [$a, $b, $c, $d, $e, $f];
}
function guardFallsThrough(int $pid) {
    $r = pcntl_waitpid($pid, $status);
    if ($r <= 0 || !pcntl_wifexited($status)) { error_log('child failed'); }
    while ($r > 0 && pcntl_wifexited($status)) { break; }
    return <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
}
