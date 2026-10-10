<?php
function sequential($o) {
    return array_values($o->Data()) === $o->data();
}
