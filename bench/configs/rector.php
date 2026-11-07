<?php

use Rector\Config\RectorConfig;

return RectorConfig::configure()
    ->withPaths([
        '@CORPUS@/@PATH@',
    ])
    ->withCache('@CACHE@/rector')
    ->withPreparedSets(deadCode: true, codeQuality: true);
