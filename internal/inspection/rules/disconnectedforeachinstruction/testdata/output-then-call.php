<?php
function progress(array $items, bool $verbose)
{
    foreach ($items as $item) {
        handle($item);
        if ($verbose) {
            echo '.';
            note($verbose);
        }
    }
}
