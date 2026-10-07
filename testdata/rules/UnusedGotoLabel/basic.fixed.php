<?php
function retry_job($job)
{
    $attempts = 0;
    again:
    if (!$job->run() && ++$attempts < 3) {
        goto again;
    }
}
