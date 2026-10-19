<?php
// A jump can land between a guard and the use: the guard proves nothing.
function jumpPastGuard() {
    if (<warning descr="Reject signal-wait failure before handling the result.">$signal = pcntl_sigwaitinfo([SIGTERM], $info)</warning>) {
        goto deliver;
        if ($signal === -1) { return; }
        deliver:
        handleSignal($signal);
    }
}
