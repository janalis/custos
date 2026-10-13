<?php
namespace Ids {
    $a = UniqId('x', true);
    $b = Array_Map(function ($value = '') { return uniqid((string) $value, true); }, ['a']);
}
