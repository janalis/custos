<?php
function buildSchema(array $fields): array {
    $schema = ['type' => 'object'];
    foreach ($fields as $field) {
        $schema['required'] ??= [];
        $schema['required'][] = $field;
    }
    $schema['tags'] = [];
    $schema['tags'][] = 'generated';
    return $schema;
}
