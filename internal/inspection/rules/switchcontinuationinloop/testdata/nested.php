<?php
foreach ($a as $v) {
    switch ($v) {
        case 1:
            switch ($v) {
                default:
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            }
    }
}
