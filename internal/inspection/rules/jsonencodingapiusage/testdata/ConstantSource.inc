<?php

namespace Lib;

const LIMIT = 10;
const LABEL = 'label';

interface Keys
{
    const PLAIN = 'callbacks';
    const QUOTED = "groups";
    const COUNT = -3;
    const RATIO = 0.5;
    const ESCAPED = 'it\'s';
    const JOINED = 'a' . 'b';
    const VARIABLE = "x\n";
}

enum Mode
{
    case On;
}
