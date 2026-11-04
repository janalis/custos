<?php
function appendRows($rows, $sink) {
    foreach ($rows as $row) {
        $sink?->append($row);
        $sink?->flush();
    }
}
