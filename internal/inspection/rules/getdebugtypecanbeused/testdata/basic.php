<?php
namespace Shop;

function label(object|string $payload, array $rows): array {
    return [
        <weak_warning descr="Use 'get_debug_type($rows[0])' instead (scalar type names differ from gettype()).">is_object($rows[0]) ? get_class($rows[0]) : gettype($rows[0])</weak_warning>,
        'kind: ' . (<weak_warning descr="Use '\get_debug_type($payload)' instead (scalar type names differ from gettype()).">\is_object($payload) ? \get_class($payload) : gettype($payload)</weak_warning>),
        is_object($payload) ? get_class($payload) : gettype($rows),
        is_object($payload) ? gettype($payload) : get_class($payload),
        is_object($payload) ? get_class($payload) : 'scalar',
        is_object($payload) ?: gettype($payload),
        (is_object($payload)) ? get_class($payload) : gettype($payload),
    ];
}
