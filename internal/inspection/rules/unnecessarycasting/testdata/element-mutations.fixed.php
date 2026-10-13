<?php
function incrementedElements() {
    $values = ['count' => 1, 'label' => 'ready'];
    ++$values['count'];
    return [
        (int) $values['count'],
        $values['label'],
    ];
}

function decrementedElements() {
    $values = ['count' => 1];
    $values['count']--;
    $copy = $values;
    return (int) $copy['count'];
}
