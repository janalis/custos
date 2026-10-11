<?php
// A jump out of 'finally' does not compile: reported without a fix.
function cleanup(array $jobs) {
    foreach ($jobs as $job) {
        try {
            run($job);
        } finally {
            switch ($job) {
                case 1:
                    continue;
            }
        }
    }
    foreach ($jobs as $job) {
        try {
            switch ($job) {
                case 1:
                    continue 2;
            }
        } catch (Exception $e) {
            switch ($job) {
                case 2:
                    continue 2;
            }
        } finally {
            foreach ($jobs as $other) {
                switch ($other) {
                    case 3:
                        continue 2;
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
                continue 2 /* keep */ ;
            case 2:
                CONTINUE 2;
        endswitch;
    endforeach;
}
