<?php
function render(array $entries)
{
    foreach ($entries as $entry) {
        $entry[2] ??= 'none';
        [$id, $label, $note] = $entry;
        echo $id, $label, $note;
    }
    foreach ($entries as $entry) {
        echo count($entry);
        [$id, $label] = $entry;
        echo $id, $label;
    }
    foreach ($entries as $entry) {
        <weak_warning descr="Destructure directly in the foreach header.">[$id, $label] = $entry</weak_warning>;
        echo $id, $label;
    }
}
