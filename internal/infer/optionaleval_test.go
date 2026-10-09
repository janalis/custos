package infer_test

import "testing"

func TestOptionalMutationEvaluation(t *testing.T) {
	checkAnywhere(t, `<?php
class OptionalMutationObject {
    public OptionalMutationObject $child;
    public array $values;
    public function run($x): self { return $this; }
}
function optionalMutations(?OptionalMutationObject $o, OptionalMutationObject $certain, array $objects) {
    $direct = null; $o?->run(++$direct); t('direct', $direct);
    $chain = null; $o?->child->run(++$chain); t('chain', $chain);
    $name = null; $o?->{++$name}; t('name', $name);
    $method = null; $o?->{++$method}(); t('method', $method);
    $offset = null; $o?->values[++$offset]; t('offset', $offset);
    $callable = null; $o?->run(1)(++$callable); t('callable', $callable);
    $receiver = null; $objects[++$receiver]?->run(1); t('receiver', $receiver);
    $nonnull = null; $certain->run(++$nonnull); t('nonnull', $nonnull);
    $left = 1; $coalesce = null; $left ??= ++$coalesce; t('coalesce', $coalesce);
    $target = null; $new = null; $target ??= ++$new; t('coalesceNullable', $new);
    $condition = null;
    if ($o?->run(++$condition)) { t('guard', $condition); }
}
`, map[string]string{
		"direct": "int|null", "chain": "int|null", "name": "int|null", "method": "int|null",
		"offset": "int|null", "callable": "int|null", "receiver": "int", "nonnull": "int",
		"coalesce": "int|null", "coalesceNullable": "int|null", "guard": "int|null",
	})
}

func TestArrowOptionalMutationEvaluation(t *testing.T) {
	checkAnywhere(t, `<?php
class ArrowMutationObject {
    public ArrowMutationObject $child;
    public array $values;
    public function run($x): self { return $this; }
}
function optionalArrows(?ArrowMutationObject $o, ArrowMutationObject $certain, array $objects) {
    $direct = null; $f = fn() => [$o?->run(++$direct), t('direct', $direct)];
    $chain = null; $f = fn() => [$o?->child->run(++$chain), t('chain', $chain)];
    $name = null; $f = fn() => [$o?->{++$name}, t('name', $name)];
    $offset = null; $f = fn() => [$o?->values[++$offset], t('offset', $offset)];
    $receiver = null; $f = fn() => [$objects[++$receiver]?->run(1), t('receiver', $receiver)];
    $nonnull = null; $f = fn() => [$certain->run(++$nonnull), t('nonnull', $nonnull)];
    $left = 1; $coalesce = null; $f = fn() => [$left ??= ++$coalesce, t('coalesce', $coalesce)];
    $target = null; $new = null; $f = fn() => [$target ??= ++$new, t('coalesceNullable', $new)];
}
`, map[string]string{
		"direct": "int|null", "chain": "int|null", "name": "int|null", "offset": "int|null",
		"receiver": "int", "nonnull": "int", "coalesce": "int|null", "coalesceNullable": "int|null",
	})
}
