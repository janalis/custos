<?php
function checked($allowed)
{
    if (!In_Array($_SERVER['HTTP_HOST'], $allowed, true)) {
        return null;
    }
    return 'admin@' . $_SERVER['HTTP_HOST'];
}
