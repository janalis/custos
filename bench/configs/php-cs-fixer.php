<?php

$finder = PhpCsFixer\Finder::create()
    ->in('@CORPUS@/@PATH@')
;

return (new PhpCsFixer\Config())
    ->setRules(['@PSR12' => true])
    ->setParallelConfig(PhpCsFixer\Runner\Parallel\ParallelConfigFactory::detect())
    ->setFinder($finder);
