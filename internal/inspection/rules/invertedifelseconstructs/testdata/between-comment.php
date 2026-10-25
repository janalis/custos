<?php
function queue($job) {
    if (!$job->ready()) {
        return 'later';
    }
    // Ready: run it now.
    <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> {
        return 'now';
    }
}

function plain($job) {
    if (!$job->ready()) {
        return 'later';
    } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> {
        // Ready: run it now.
        return 'now';
    }
}
