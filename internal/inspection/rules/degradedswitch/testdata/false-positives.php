<?php
function route($verb) {
    switch ($verb) {}
    switch ($verb) {
        case 'PUT':
        case 'PATCH':
            update();
    }
    switch ($verb) {
        case 'A': a(); break;
        case 'B': b(); break;
        default: c();
    }
}
