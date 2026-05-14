package hyperloglog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSketchPool_GetPut(t *testing.T) {
	pool := NewSketchPool(14, true)

	sk, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.Equal(t, uint8(14), sk.p)
	require.True(t, sk.s)

	require.NoError(t, pool.Put(sk))
}

func TestSketchPool_PutNil(t *testing.T) {
	pool := NewSketchPool(14, true)
	require.NoError(t, pool.Put(nil))
}

func TestSketchPool_PutWrongPrecision(t *testing.T) {
	pool := NewSketchPool(14, true)

	sk, err := NewSketch(10, true)
	require.NoError(t, err)

	err = pool.Put(sk)
	require.Error(t, err)
}

func TestSketchPool_ReuseResetsSketch(t *testing.T) {
	pool := NewSketchPool(14, true)

	sk, err := pool.Get()
	require.NoError(t, err)

	for i := uint64(0); i < 1000; i++ {
		sk.InsertHash(i)
	}
	require.Greater(t, sk.Estimate(), uint64(0))

	require.NoError(t, pool.Put(sk))

	sk2, err := pool.Get()
	require.NoError(t, err)
	// May or may not be the same object, but must be reset.
	require.Equal(t, uint64(0), sk2.Estimate())
}

func TestSketchPool_MustGetPut(t *testing.T) {
	pool := NewSketchPool(14, false)

	sk := pool.MustGet()
	require.NotNil(t, sk)
	require.Equal(t, uint8(14), sk.p)
	require.False(t, sk.s)

	pool.MustPut(sk)
}

func TestSketchPoolPool_GetPutSparse(t *testing.T) {
	skpp := NewSketchPoolPool()

	sk, err := skpp.Get(14, true)
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.Equal(t, uint8(14), sk.p)
	require.True(t, sk.s)

	require.NoError(t, skpp.Put(sk))
}

func TestSketchPoolPool_GetPutNormal(t *testing.T) {
	skpp := NewSketchPoolPool()

	sk, err := skpp.Get(14, false)
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.Equal(t, uint8(14), sk.p)
	require.False(t, sk.s)

	require.NoError(t, skpp.Put(sk))
}

func TestSketchPoolPool_PutNil(t *testing.T) {
	skpp := NewSketchPoolPool()
	require.NoError(t, skpp.Put(nil))
	skpp.MustPut(nil) // must not panic
}

// Sparse and normal pools are separate — a sparse sketch must not land in the normal pool.
func TestSketchPoolPool_SparseAndNormalAreIsolated(t *testing.T) {
	skpp := NewSketchPoolPool()

	sparse := skpp.MustGet(14, true)
	normal := skpp.MustGet(14, false)

	skpp.MustPut(sparse)
	skpp.MustPut(normal)

	// Getting again should return sketches with the correct s flag.
	sk1 := skpp.MustGet(14, true)
	require.True(t, sk1.s)

	sk2 := skpp.MustGet(14, false)
	require.False(t, sk2.s)
}

// A sparse sketch that converted to normal internally is still returned to the sparse pool.
func TestSketchPoolPool_SparseConvertedToNormalRoutedCorrectly(t *testing.T) {
	skpp := NewSketchPoolPool()

	sk := skpp.MustGet(4, true)
	require.True(t, sk.s)

	// Drive it past the sparse→normal threshold.
	for i := uint64(0); i < 100_000; i++ {
		sk.InsertHash(i)
	}
	require.False(t, sk.sparse(), "expected sketch to have converted to normal mode")

	// Put should route it back to the sparse pool (sk.s is still true).
	skpp.MustPut(sk)

	// Retrieve from sparse pool — must succeed and be reset.
	recovered := skpp.MustGet(4, true)
	require.Equal(t, uint64(0), recovered.Estimate())
}

func TestSketchPoolPool_AllPrecisions(t *testing.T) {
	skpp := NewSketchPoolPool()

	for p := uint8(4); p <= 18; p++ {
		for _, sparse := range []bool{true, false} {
			sk, err := skpp.Get(p, sparse)
			require.NoError(t, err, "precision=%d sparse=%v", p, sparse)
			require.Equal(t, p, sk.p)
			require.Equal(t, sparse, sk.s)
			require.NoError(t, skpp.Put(sk))
		}
	}
}
