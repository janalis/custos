<?php

namespace App;

use JetBrains\PhpStorm\ExpectedValues;

const LOCAL = 1;

function local(#[ExpectedValues([LOCAL])] int $value) {}

function fallsBack(#[ExpectedValues([E_ERROR])] int $value) {}

local(<warning descr="Pass App\LOCAL to $value of App\local(), not 2.">2</warning>);
local(1);
fallsBack(<warning descr="Pass App\E_ERROR to $value of App\fallsBack(), not 2.">2</warning>);
fallsBack(1);
