<?php
// Statements after the switch run after 'continue' but not after
// 'continue 2': reported without a fix.
function tokens(array $parts) {
    $i = 0;
    foreach ($parts as $part) {
        switch ($part) {
            case ',':
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
        }
        ++$i;
    }
    while ($i-- > 0) {
        if ($i > 3) {
            switch ($i) {
                case 5:
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
        }
        echo $i;
    }
    for ($j = 0; $j < 3; $j++) {
        switch ($j) {
            case 1:
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
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
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
        }
    }
}
