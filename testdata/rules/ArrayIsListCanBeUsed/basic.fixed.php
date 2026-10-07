<?php
function inspect(array $cfg, array $other) {
    return [
        array_is_list($cfg),
        !array_is_list($cfg),
        !\array_is_list($cfg),
        array_is_list($cfg),
        !array_is_list($cfg),
        array_is_list($this->rows['a']),
    ];
}
