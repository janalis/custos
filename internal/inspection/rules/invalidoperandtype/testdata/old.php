<?php

function beforeTypeErrors(array $row)
{
    return $row * 2;
}
