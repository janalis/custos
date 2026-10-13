<?php
namespace Api;

function json_encode($v) { return ''; }

function out(array $rows) {
    return [
        json_encode($rows),
        \Json_Encode($rows, JSON_THROW_ON_ERROR),
        JSON_DECODE($rows[0], true, 512, JSON_THROW_ON_ERROR),
    ];
}
