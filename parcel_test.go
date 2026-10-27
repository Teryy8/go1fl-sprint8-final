package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	assert.NotEqual(t, id, 0)
	parcel.Number = id

	// get
	testerGet, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, parcel.Number, testerGet.Number)
	assert.Equal(t, parcel.Client, testerGet.Client)
	assert.Equal(t, parcel.Status, testerGet.Status)
	assert.Equal(t, parcel.Address, testerGet.Address)
	assert.Equal(t, parcel.CreatedAt, testerGet.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	_ , err = store.Get(id)
	assert.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	assert.NotEqual(t, id, 0)

	// set address
	
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)

	require.NoError(t, err)

	// check
	testerGet, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, newAddress, testerGet.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	assert.NotEqual(t, id, 0)

	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err)

	// check
	testerGet, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, ParcelStatusDelivered, testerGet.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	store := NewParcelStore(db)
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		assert.NotEqual(t, id, 0)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	assert.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
