<?php
function incrementedElements() {
    $values = ['count' => 1, 'label' => 'ready'];
    ++$values['count'];
    return [
        (int) $values['count'],
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $values['label'],
    ];
}

function decrementedElements() {
    $values = ['count' => 1];
    $values['count']--;
    $copy = $values;
    return (int) $copy['count'];
}
