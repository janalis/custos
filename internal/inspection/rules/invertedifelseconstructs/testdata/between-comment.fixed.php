<?php
function queue($job) {
    if (!$job->ready()) {
        return 'later';
    }
    // Ready: run it now.
    else {
        return 'now';
    }
}

function plain($job) {
    if ($job->ready()) {
        // Ready: run it now.
        return 'now';
    } else {
        return 'later';
    }
}
