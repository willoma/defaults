package defaults_test

import (
	"io/fs"
	"maps"
	"math/big"
	"net"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/willoma/defaults"
)

type unmarshaltarget struct {
	A time.Time        `default:"1982-04-12T23:20:00+02:00"`
	B time.Time        `default:"1982-04-12"`
	C time.Time        `default:"23:20"`
	D map[string]int   `default:"foo:1,bar:2"`
	E chan int         `default:"5"`
	F *net.IP          `default:"192.168.42.2"`
	G string           `default:"alice, bob"`
	H *string          `default:"some default"`
	I big.Int          `default:"12345789123456789123456789123456789123456789"`
	J []uint8          `default:"2,8"`
	K []string         `default:"foo,bar,b\\,az"`
	L net.HardwareAddr `default:"aa:bb:cc:dd:ee:ff"`
	M net.IP           `default:"192.168.42.1"`
	N big.Float        `default:"3.14159265358979323846264338327950288419716939937510582097494459230781640628"`
	O regexp.Regexp    `default:"(foo|bar)"`
	P struct {
		Q bool `default:"true"`
		R uint `default:"42"`
		S int  `default:"42"`
	}
	T float64       `default:"3.14159"`
	U time.Duration `default:"42s"`
	V fs.FileMode   `default:"754"`
	W fs.FileMode   `default:"rwxr-xr-x"`
	X int16         `default:"42"`
	Y [3]bool       `default:"true,false,true"`
	Z bool          `default:"t"`
}

func TestApply(t *testing.T) {
	t.Parallel()

	value := unmarshaltarget{}
	if err := defaults.Complete(&value); err != nil {
		t.Fatalf("failed to apply defaults: %s", err)
	}

	if !value.A.Equal(time.Date(1982, 4, 12, 21, 20, 0, 0, time.UTC)) {
		t.Errorf("wrong value for A: %q", value.A)
	}

	if !value.B.Equal(time.Date(1982, 4, 12, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("wrong value for B: %q", value.B)
	}

	if !value.C.Equal(time.Date(0, 1, 1, 23, 20, 0, 0, time.UTC)) {
		t.Errorf("wrong value for C: %q", value.C)
	}

	if !maps.Equal(value.D, map[string]int{"foo": 1, "bar": 2}) {
		t.Errorf("wrong value for D: %v", value.D)
	}

	if cap(value.E) != 5 {
		t.Errorf("wrong chan buffer size for E: %d", cap(value.E))
	}

	select {
	case value.E <- 42:
	default:
		t.Error("failed to send to chan E")
	}

	select {
	case r := <-value.E:
		if r != 42 {
			t.Errorf("wrong value in E: %d", r)
		}

	default:
		t.Error("failed to receive from chan E")
	}

	if value.F == nil {
		t.Error("F is nil")
	} else if !value.F.Equal(net.ParseIP("192.168.42.2")) {
		t.Errorf("wrong value for F: %q", value.F.String())
	}

	if value.G != "alice, bob" {
		t.Errorf("wrong value for G: %q", value.G)
	}

	if value.H == nil {
		t.Error("H is nil")
	} else if *value.H != "some default" {
		t.Errorf("wrong value for H: %q", *value.H)
	}

	var bigInt big.Int
	if err := bigInt.UnmarshalText([]byte("12345789123456789123456789123456789123456789")); err != nil {
		t.Fatal(err)
	}

	if value.I.Cmp(&bigInt) != 0 {
		t.Errorf("wrong value for I: %q", value.I.String())
	}

	if !slices.Equal(value.J, []uint8{2, 8}) {
		t.Errorf("wrong value for J: %q", value.J)
	}

	if !slices.Equal(value.K, []string{"foo", "bar", "b,az"}) {
		t.Errorf("wrong value for K: %q", value.K)
	}

	if value.L.String() != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("wrong value for L: %q", value.L.String())
	}

	if !value.M.Equal(net.ParseIP("192.168.42.1")) {
		t.Errorf("wrong value for M: %q", value.M.String())
	}

	var pi314 big.Float
	if err := pi314.UnmarshalText(
		[]byte("3.14159265358979323846264338327950288419716939937510582097494459230781640628"),
	); err != nil {
		t.Fatal(err)
	}

	if value.N.Cmp(&pi314) != 0 {
		t.Errorf("wrong value for N: %q", value.N.String())
	}

	if value.O.String() != "(foo|bar)" {
		t.Errorf("wrong value for O: %q", value.O.String())
	}

	if !value.P.Q {
		t.Errorf("wrong value for P.Q: %t", value.P.Q)
	}

	if value.P.R != 42 {
		t.Errorf("wrong value for P.R: %d", value.P.R)
	}

	if value.P.S != 42 {
		t.Errorf("wrong value for P.S: %d", value.P.S)
	}

	if value.T != 3.14159 {
		t.Errorf("wrong value for T: %f", value.T)
	}

	if value.U != 42*time.Second {
		t.Errorf("wrong value for U: %q", value.U)
	}

	if value.V != fs.FileMode(0o754) {
		t.Errorf("wrong value for V: %q", value.V)
	}

	if value.W != fs.FileMode(0o755) {
		t.Errorf("wrong value for W: %q", value.W)
	}

	if value.X != 42 {
		t.Errorf("wrong value for X: %d", value.X)
	}

	if value.Y != [3]bool{true, false, true} {
		t.Errorf("wrong value for Y: %v", value.Y)
	}

	if !value.Z {
		t.Errorf("wrong value for Z: %t", value.Z)
	}
}
