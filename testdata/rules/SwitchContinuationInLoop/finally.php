<?php
// A jump out of 'finally' does not compile: reported without a fix.
function cleanup(array $jobs) {
    foreach ($jobs as $job) {
        try {
            run($job);
        } finally {
            switch ($job) {
                case 1:
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
        }
    }
    foreach ($jobs as $job) {
        try {
            switch ($job) {
                case 1:
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
        } catch (Exception $e) {
            switch ($job) {
                case 2:
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
        } finally {
            foreach ($jobs as $other) {
                switch ($other) {
                    case 3:
                        <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
                }
            }
        }
    }
}

// Only the level is inserted: comments and case of the keyword stay.
function keep(array $jobs) {
    foreach ($jobs as $job):
        switch ($job):
            case 1:
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue /* keep */ ;</error>
            case 2:
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">CONTINUE;</error>
        endswitch;
    endforeach;
}
