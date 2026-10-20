package infer_test

import "testing"

func TestRound19ArithmeticAndWrites(t *testing.T) {
	check(t, `<?php
function overwrite(int &$n) { $n = 'changed'; }
function run(int|float $n, float $f, int $i, bool $b, bool|string $bs, $unknown, array $plain) {
 t('union',$n+1); t('float',$f+1); t('int',$i+1); t('unknown',$unknown+1);
 t('small',2**3); t('overflow',2**63); t('negative',2**-1); t('floating',2.0**2); t('powerUnknown',$unknown**2); t('variable',$i**2); t('ternary',$b?:null); t('ternaryUnion',$bs?:null); t('falseTernary',false?:1); t('unknownTernary',$unknown?:1);
 $n2=2; $n2**=2; t('compound',$n2);
 $v=1; overwrite($v); t('overwritten',$v);
 $a=['a'=>1]; sort($a); t('sorted',$a);
 $a2=['a'=>1]; ksort($a2); t('keysorted',$a2);
 $empty=[]; array_push($empty,1); t('emptyPush',$empty);
 $a3=[1]; array_push($a3,...['s']); t('pushed',$a3);
 $a4=[1]; array_unshift($a4,'s'); t('unshifted',$a4);
 $a5=[1]; array_splice($a5,0,1,['s']); t('spliced',$a5);
 sort($plain); t('plain',$plain);
 $s='a'; settype($s,'int'); t('settype',$s);
 $b2='a'; settype($b2,'bool'); t('setbool',$b2);
 $f2='a'; settype($f2,'double'); t('setfloat',$f2);
 $s2=1; settype($s2,'string'); t('setstring',$s2);
 $a6=1; settype($a6,'array'); t('setarray',$a6);
 $o=1; settype($o,'object'); t('setobject',$o);
 $null=1; settype($null,'null'); t('setnull',$null);
 $bad=1; settype($bad,'bad'); t('bad',$bad);
 $dynamic=1; settype($dynamic,$unknown); t('dynamic',$dynamic);
}
`, map[string]string{
		"union": "float|int", "float": "float", "int": "int", "unknown": "float|int", "small": "int", "overflow": "float|int", "negative": "float|int", "floating": "float|int", "powerUnknown": "float|int", "variable": "float|int", "ternary": "null|true", "ternaryUnion": "null|string|true", "falseTernary": "int", "unknownTernary": "?unknown", "compound": "float|int", "overwritten": "?unknown",
		"sorted": "int[]", "keysorted": "int[]", "pushed": "int[]|string[]", "emptyPush": "array", "unshifted": "int[]|string[]", "spliced": "int[]|string[]", "plain": "array", "settype": "int", "setbool": "bool", "setfloat": "float", "setstring": "string", "setarray": "array", "setobject": "object", "setnull": "null", "bad": "?unknown", "dynamic": "?unknown",
	})
}

func TestRound19TemplateUnionAndCountable(t *testing.T) {
	check(t, `<?php
class Model {}
/**
 * @template T
 * @param T|class-string<T> $value
 * @return T
 */
function instance($value) { return new $value; }
/**
 * @template T
 * @param T|T[] $value
 * @return T
 */
function first($value) { return $value; }
function run(mixed $value) {
 t('class',instance(Model::class)); t('one',first(1)); t('many',first([1]));
 if (is_countable($value)) { t('countable',$value); }
}
`, map[string]string{"class": `\Model`, "one": "int", "many": "int", "countable": `\Countable|array`})
}

func TestRound19NamedMutationArguments(t *testing.T) {
	check(t, `<?php
class A { function mutate(int &$v) {} }
class B { function mutate($v) {} }
function run(A $receiver) {
 $v=1; settype(type:'string',var:$v); t('namedtype',$v);
 $a=['x'=>1]; sort(flags:SORT_NUMERIC,array:$a); t('sort',$a);
 $b=[1]; array_splice(replacement:'s',length:1,offset:0,array:$b); t('scalarsplice',$b);
 $c=[1]; array_splice($c,0,1,null); t('nullsplice',$c);
 $d=['x'=>1]; array_unshift($d,'s'); t('unshift',$d);
 $e=[1]; array_push(array:$e,values:'s'); t('namedpush',$e);
 $v2=1; $receiver=new B; $receiver->mutate($v2); t('reassigned',$v2);
 $v3=1; settype($v3); t('missing',$v3);
}
`, map[string]string{"namedtype": "string", "sort": "int[]", "scalarsplice": "int[]|string[]", "nullsplice": "int[]", "unshift": "int[]|string[]", "namedpush": "int[]|string[]", "reassigned": "int", "missing": "?unknown"})
}

func TestRound19IndexedDocTypes(t *testing.T) {
	checkWith(t, map[string]string{"types.php": `<?php namespace External; enum Status {case Active;} class K {const int NUMBER=1;}`}, `<?php
use External\Status;
use External\K;
function run() {
 /** @var Status::Active|null $state */
 $state=unknown(); t('state',$state);
 /** @var K::NUMBER $number */
 $number=unknown(); t('number',$number);
}
`, map[string]string{"state": `\External\Status|null`, "number": "int"})
}

func TestRound19DeferredMethodWrites(t *testing.T) {
	check(t, `<?php
class RefService {
 function mutate(int &$value) { $value='changed'; }
 /** @param-out string $value */
 function fill(&$value) {}
 function byValue($value) {}
 /** @param-out mixed $value */
 function mixedOut(&$value) {}
}
class Owner {
 public RefService $service;
 function run($unknown,bool $condition) {
  $service=new RefService;
  $a=1; $service->mutate($a); t('local',$a);
  $b=1; $this->service->mutate($b); t('property',$b);
  $c=1; $service->fill($c); t('out',$c);
  $d=1; $service->byValue($d); t('value',$d);
  $e=1; $unknown->mutate($e); t('unresolved',$e);
  $mixed=1; $service->mixedOut($mixed); t('mixedOut',$mixed);
 $f=1; $service->missing($f); t('missing',$f);
  $g=1; $service->mutate(1,$g); t('extra',$g);
  $h=1; $service->byValue(1,...[$h]); t('unpack',$h);
  $i=1; $service=unknown(); $service->mutate($i); t('reboundUnknown',$i);
 }
}
$o=unknown(); $a=1; $o->method($a); t('file',$a);
`, map[string]string{"local": "?unknown", "property": "?unknown", "out": "string", "value": "int", "unresolved": "int", "missing": "int", "extra": "int", "unpack": "int", "reboundUnknown": "int", "mixedOut": "?unknown", "file": "int"})
}

func TestDeferredArrowAndUnpacking(t *testing.T) {
	checkAnywhere(t, `<?php
class Builder {function mutate(&$value) {}}
function run() {
 $service=new Builder; $a=1;$b=2;
 $service->missing($a,...[$b]);t('unpacked',$a);
}
$arrow=fn()=>[$service=new Builder,$value=1,$service->mutate($value),t('arrowWrite',$value)];
`, map[string]string{"unpacked": "int", "arrowWrite": "?unknown"})
}

func TestRound19DeferredIndexedContracts(t *testing.T) {
	checkWith(t, map[string]string{
		"types.php": `<?php namespace Model; const PREFIX='active';enum State:string {case Active=PREFIX;} enum Pure {case Active;} class Constants {const int CODE=1;}`,
		"service.php": `<?php namespace Api;
  use Model\State;use Model\Constants;
  /** @return State::Active|null */ function state(){}
  /** @return State::Active[] */ function states(){}
 /** @param-out State::Active $value */ function fillState(&$value){}
 /** @return value-of<State> */ function stateValue(){}
 /** @return key-of<State> */ function stateKey(){}
 /** @return value-of<\Model\Pure> */ function pureValue(){}
  /** @return Missing::Case|null */ function missing(){}
  class Service {
   /** @var State::Active|null */ public $state;
   /** @return Constants::CODE */ function code(){}
 /** @param-out State::Active $value */ function fillState(&$value){}
  }`,
	}, `<?php
 function run(\Api\Service $service){
  $output=1; \Api\fillState($output);t('directOut',$output);
  $methodOutput=1;$service->fillState($methodOutput);t('methodOut',$methodOutput);
  t('pure',\Api\pureValue());t('value',\Api\stateValue());t('key',\Api\stateKey());t('function',\Api\state());t('array',\Api\states());t('property',$service->state);t('method',$service->code());t('missing',\Api\missing());
 }
 `, map[string]string{"function": `\Model\State|null`, "value": "string", "key": "string", "pure": "mixed", "directOut": `\Model\State`, "methodOut": `\Model\State`, "array": `\Model\State[]`, "property": `\Model\State|null`, "method": "int", "missing": "mixed|null"})
}

func TestDeferredMethodOutputBranches(t *testing.T) {
	check(t, `<?php
class Writer {/** @param-out string $value */ function fill(&$value){} function next():Writer{return $this;}}
class Reader {function fill($value){}}
class Other {}
function run(Writer|Reader $union,Writer|Other $missing,?Writer $optional) {
 $a=1;$union->fill($a);t('byValueBranch',$a);
 $b=1;$missing->fill($b);t('missingBranch',$b);
 $c=1;$optional?->fill($c);t('nullable',$c);
 $d=1;$optional?->next()->fill($d);t('chain',$d);
}
`, map[string]string{"byValueBranch": "int|string", "missingBranch": "int|string", "nullable": "int|string", "chain": "int|string"})
}
