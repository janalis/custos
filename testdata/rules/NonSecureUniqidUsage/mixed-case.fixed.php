<?php
namespace Ids {
    $a = UniqId('x', true);
    $b = Array_Map(function ($value) { return uniqid($value, true); }, ['a']);
}
