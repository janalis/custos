<?php
define('EXPORT_FLAGS', JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_UNESCAPED_SLASHES);

class Exporter {
    const MODE = JSON_THROW_ON_ERROR | JSON_PRETTY_PRINT;

    function run(array $rows, int $mode, int $max) {
        $strict = JSON_THROW_ON_ERROR;
        return [
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[0])</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[1], true, $max)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[2], null, 16, JSON_BIGINT_AS_STRING)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[3], flags: $mode)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, $mode | JSON_HEX_TAG)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, $mode, $max)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, flags: $mode)</weak_warning>,

            json_decode($rows[0], true, 32, JSON_THROW_ON_ERROR),
            json_decode($rows[0], true, 32, $mode | JSON_PARTIAL_OUTPUT_ON_ERROR),
            json_decode($rows[0], flags: $strict),
            json_encode($rows, EXPORT_FLAGS),
            json_encode($rows, self::MODE),
            json_encode($rows, 4194304),
            json_encode(),
        ];
    }
}
