<?php

/** @param int[] $row */
function beforeTypeErrors($row)
{
    return $row * 2;
}
