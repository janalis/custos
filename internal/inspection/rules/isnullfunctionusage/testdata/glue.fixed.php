<?php
function glued($row, $ok)
{
    if ($ok) {
        return load() !== null;
    }
    if ($row) {
        return \App\load() !== null;
    }
    echo 1 !== null;
    $a = $row === null and$ok;
    $b = $row === null or $ok;
    return$row !== null;
}
