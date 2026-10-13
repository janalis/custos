<?php
define('EXPORT_FLAGS', JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_UNESCAPED_SLASHES);

class Exporter {
    const MODE = JSON_THROW_ON_ERROR | JSON_PRETTY_PRINT;

    function run(array $rows, int $mode, int $max) {
        $strict = JSON_THROW_ON_ERROR;
        return [
            json_decode($rows[0], false, 512, JSON_THROW_ON_ERROR),
            json_decode($rows[1], true, $max, JSON_THROW_ON_ERROR),
            json_decode($rows[2], null, 16, JSON_THROW_ON_ERROR | JSON_BIGINT_AS_STRING),
            json_decode($rows[3], flags: JSON_THROW_ON_ERROR | $mode),
            json_encode($rows, JSON_THROW_ON_ERROR),
            json_encode($rows, JSON_THROW_ON_ERROR | $mode | JSON_HEX_TAG),
            json_encode($rows, JSON_THROW_ON_ERROR | $mode, $max),
            json_encode($rows, flags: JSON_THROW_ON_ERROR | $mode),

            json_decode($rows[0], true, 32, JSON_THROW_ON_ERROR),
            json_decode($rows[0], true, 32, $mode | JSON_PARTIAL_OUTPUT_ON_ERROR),
            json_decode($rows[0], flags: $strict),
            json_encode($rows, EXPORT_FLAGS),
            json_encode($rows, self::MODE),
            json_encode($rows, 4194304),
            json_encode(),
            json_encode(...$rows),
            json_encode($rows, -4194304),
        ];
    }
}
