package catchcannothandleknownthrowable

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CatchCannotHandleKnownThrowable"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\ntry { throw new Error('failed'); } catch (Exception $e) {}\n", 1},
		{"\ntry { throw new Error('failed'); } catch (Throwable $e) {}\n", 0},
		{"try{throw new TypeError('x');}catch(Exception|RuntimeException $e){}", 1},
		{"try{throw new Exception('x');}catch(Error $e){}", 0},
		{"try{throw new Error('x');}catch(Missing $e){}", 0},
		{"try{unknown();throw new Error('x');}catch(Exception $e){}", 0},
		{"try{throw new Error('x');}catch(Exception $e){}finally{}", 0},
		{"try{$x=1;}catch(Exception $e){}", 0},
		{"try{throw $unknown;}catch(Exception $e){}", 0},
		{"class X extends Error{function __construct(){}}try{throw new X();}catch(Exception $e){}", 0},
		{"try{throw new Error(unknown());}catch(Exception $e){}", 0},
		{"try{throw new Error('x');}catch(Error $e){}catch(Exception $e){}", 0},
		{"if(false){\ntry { throw new Error('failed'); } catch (Exception $e) {}\n}", 0},
		{"\ntry { throw new Error('failed'); } catch (Exception $e) {}\n $broken = ;", 0},
		{"try{return;}catch(Exception $e){}", 0},
		{"try{throw new Error;}catch(Exception $e){}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
