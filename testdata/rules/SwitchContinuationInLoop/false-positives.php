<?php
switch ($x) {
    case 1:
        continue;
}
foreach ($a as $v) {
    switch ($v) {
        case 1:
            continue 1;
        case 2:
            foreach ($v as $w) { continue; }
    }
    continue;
}
function f() {
    switch (1) { default: continue; }
}
