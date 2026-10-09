package infer_test

import (
	"testing"

	phpversion "custos/internal/php/version"
)

// Magic signatures must reach the same mutation analysis as real methods.
func TestMagicMethodReferenceParametersInvalidateShapes(t *testing.T) {
	got := shapesAt(t, phpversion.PHP84, `<?php
namespace App;
/**
 * @method void inspect(array $values)
 * @method void mutate(array &$values)
 * @method static void rewrite(array &$values)
 * @method void many(array &...$values)
 */
class Service {}
function run(Service $service) {
    $plain = ['k' => 1];
    $service->inspect($plain);
    t('plain', $plain);
    $instance = ['k' => 1];
    $service->mutate($instance);
    t('instance', $instance);
    $named = ['k' => 1];
    $service->mutate(values: $named);
    t('named', $named);
    $static = ['k' => 1];
    Service::rewrite($static);
    t('static', $static);
    $variadic = ['k' => 1];
    $service->many([], $variadic);
    t('variadic', $variadic);
}
`)
	expectShapes(t, got, map[string]string{
		"plain":    "int[]{k: int}",
		"instance": "int[]",
		"named":    "int[]",
		"static":   "int[]",
		"variadic": "int[]",
	})
}
