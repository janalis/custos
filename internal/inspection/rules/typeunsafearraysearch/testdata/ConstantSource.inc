<?php

namespace Lib;

interface Kinds
{
    const PLAIN = 'plain';
    const QUOTED = "quoted";
    const NUMERIC = '10';
    const ESCAPED = 'it\'s';
    const JOINED = 'a' . 'b';
    const VARIABLE = "x\n";
}
