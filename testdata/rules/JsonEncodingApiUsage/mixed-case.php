<?php
namespace Api;

function json_encode($v) { return ''; }

function out(array $rows) {
    return [
        json_encode($rows),
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\Json_Encode($rows)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">JSON_DECODE($rows[0], true)</weak_warning>,
    ];
}
