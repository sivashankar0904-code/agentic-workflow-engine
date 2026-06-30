import io
import requests
import streamlit as st
import pandas as pd
import pdfplumber

API_URL = "http://localhost:8001/api/v1/chat/"

st.title("Mock Service 1")

file = st.file_uploader("Upload a file", type=["csv", "xlsx", "xls", "pdf"])

message = None

if file is not None:
    ext = file.name.rsplit(".", 1)[-1].lower()

    if ext == "csv":
        df = pd.read_csv(file)
        st.dataframe(df)
        message = df.to_string(index=False)

    elif ext in ("xlsx", "xls"):
        df = pd.read_excel(file)
        st.dataframe(df)
        message = df.to_string(index=False)

    elif ext == "pdf":
        with pdfplumber.open(io.BytesIO(file.read())) as pdf:
            text = "\n".join(page.extract_text() or "" for page in pdf.pages)
        st.text_area("Extracted Text", text, height=300)
        message = text

if st.button("Send") and message:
    try:
        response = requests.post(API_URL, json={"message": message})
        response.raise_for_status()
        st.success(response.json()["reply"])
    except requests.exceptions.ConnectionError:
        st.error("Could not connect to the API. Make sure mock_service_1 is running on port 8001.")
    except Exception as e:
        st.error(f"Error: {e}")
