<?php
function retry_job($job)
{
    $attempts = 0;
    again:
    <weak_warning descr="Label 'done' is never targeted by a goto; remove it.">done:</weak_warning>
    <weak_warning descr="Label 'Again' is never targeted by a goto; remove it.">Again:</weak_warning>
    if (!$job->run() && ++$attempts < 3) {
        goto again;
    }
}
