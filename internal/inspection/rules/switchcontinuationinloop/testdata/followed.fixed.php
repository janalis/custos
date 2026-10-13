<?php
// Statements after the switch run after 'continue' but not after
// 'continue 2': reported without a fix.
function tokens(array $parts) {
    $i = 0;
    foreach ($parts as $part) {
        switch ($part) {
            case ',':
                continue;
        }
        ++$i;
    }
    while ($i-- > 0) {
        if ($i > 3) {
            switch ($i) {
                case 5:
                    continue;
            }
        }
        echo $i;
    }
    for ($j = 0; $j < 3; $j++) {
        switch ($j) {
            case 1:
                continue 2;
        }
    }
    return $i;
}

function branches(array $parts) {
    foreach ($parts as $part) {
        if ($part === '') {
            echo 'empty';
        } else {
            switch ($part) {
                case ',':
                    continue 2;
            }
        }
    }
}
