<?php
function sync(array $batches, array $retries)
{
    foreach ($batches as $batch) {
        foreach ($batch as $line) {
            store($line);
        }
        report($line);                  // last $line of this batch: connected
    }

    foreach ($batches as $batch) {
        try {
            push($batch);
        } catch (\RuntimeException $error) {
        }
        log_error($error);              // caught in this iteration: connected
    }

    foreach ($batches as $batch) {
        <weak_warning descr="Statement does not depend on the loop; move it out.">foreach</weak_warning> ($retries as [$code, $delay]) {
            wait_for($delay);
        }
        notify($code);                  // destructured header: connected
        handle($batch);
    }

    foreach ($batches as $batch) {
        <weak_warning descr="Statement does not depend on the loop; move it out.">foreach</weak_warning> ($retries as $retry) {
            wait_for($retry);
        }
        <weak_warning descr="Statement does not depend on the loop; move it out.">try</weak_warning> {
            ping();
        } catch (\Exception $e) {
            report($e);
        }
        handle($batch);
    }
}
