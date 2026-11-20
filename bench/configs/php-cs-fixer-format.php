<?php

// Layout only: the PER Coding Style set, without risky rules.
$finder = PhpCsFixer\Finder::create()
    ->in('@CORPUS@/@PATH@')
;

return (new PhpCsFixer\Config())
    ->setRules(['@PER-CS' => true])
    ->setParallelConfig(PhpCsFixer\Runner\Parallel\ParallelConfigFactory::detect())
    ->setFinder($finder);
