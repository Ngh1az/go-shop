import { useEffect, useState } from 'react'
import axios from 'axios'
import './App.css'

const emptyForm = { id: null, name: '', price: '', description: '' }

function App() {
  const [products, setProducts] = useState([])
  const [form, setForm] = useState(emptyForm)
  const [imageFile, setImageFile] = useState(null)
  const [error, setError] = useState('')

  async function fetchProducts() {
    try {
      const res = await axios.get('/api/products')
      setProducts(res.data)
      setError('')
    } catch (err) {
      setError(err.response?.data?.message || 'Không tải được danh sách sản phẩm')
    }
  }

  useEffect(() => {
    fetchProducts()
  }, [])

  function resetForm() {
    setForm(emptyForm)
    setImageFile(null)
  }

  async function handleSubmit(e) {
    e.preventDefault()

    const data = new FormData()
    data.append('name', form.name)
    data.append('price', form.price)
    data.append('description', form.description)
    if (imageFile) data.append('image', imageFile)

    try {
      if (form.id) {
        await axios.put(`/api/products/${form.id}`, data)
      } else {
        await axios.post('/api/products', data)
      }
      resetForm()
      await fetchProducts()
    } catch (err) {
      setError(err.response?.data?.message || 'Không lưu được sản phẩm')
    }
  }

  function handleEdit(product) {
    setForm({
      id: product.id,
      name: product.name,
      price: product.price,
      description: product.description,
    })
    setImageFile(null)
  }

  async function handleDelete(id) {
    if (!window.confirm('Xoá sản phẩm này?')) return

    try {
      await axios.delete(`/api/products/${id}`)
      await fetchProducts()
    } catch (err) {
      setError(err.response?.data?.message || 'Không xoá được sản phẩm')
    }
  }

  return (
    <div className="container">
      <h1>Go Shop - Quản lý sản phẩm</h1>

      {error && <div className="error">{error}</div>}

      <form className="product-form" onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Tên sản phẩm"
          value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })}
          required
        />
        <input
          type="number"
          placeholder="Giá"
          value={form.price}
          onChange={(e) => setForm({ ...form, price: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Mô tả"
          value={form.description}
          onChange={(e) => setForm({ ...form, description: e.target.value })}
        />
        <input
          type="file"
          accept="image/*"
          onChange={(e) => setImageFile(e.target.files[0])}
        />
        <div className="form-actions">
          <button type="submit">{form.id ? 'Cập nhật' : 'Thêm sản phẩm'}</button>
          {form.id && (
            <button type="button" onClick={resetForm}>
              Huỷ
            </button>
          )}
        </div>
      </form>

      <div className="product-grid">
        {products.map((p) => (
          <div className="product-card" key={p.id}>
            {p.image && <img src={p.image} alt={p.name} />}
            <h3>{p.name}</h3>
            <p className="price">{p.price.toLocaleString('vi-VN')} đ</p>
            {p.description && <p className="description">{p.description}</p>}
            <div className="card-actions">
              <button onClick={() => handleEdit(p)}>Sửa</button>
              <button onClick={() => handleDelete(p.id)}>Xoá</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default App
